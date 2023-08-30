package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/alecthomas/kong"
	"knot8.io/pkg/lensed"
)

// Context is a CLI context.
type Context struct {
	*CLI
}

// CLI contains the CLI parameters.
type CLI struct {
	Set SetCmd `cmd:""`
	Get GetCmd `cmd:""`

	Version kong.VersionFlag `name:"version" help:"Print version information and quit"`
}

type CommonFlags struct {
	Path string `name:"filename" short:"f" optional:"" help:"Filename of the file to operate on" type:"file"`
}

type SetCmd struct {
	CommonFlags
	Values []Setter `arg:"" required:"" help:"Value to set. Format: pointer=value or pointer=@filename, where a leading @ can be escaped with a backslash."`
}

func (cmd *SetCmd) Run(cli *Context) error {
	f, err := newShadowFile(cmd.Path)
	if err != nil {
		return err
	}

	var mappings []lensed.Mapping
	for _, v := range cmd.Values {
		mappings = append(mappings, lensed.Mapping{
			Pointer:     v.Field,
			Replacement: v.Value,
		})
	}

	b, err := lensed.Apply(f.buf, mappings)
	if err != nil {
		return err
	}
	f.buf = b
	return f.Commit()
}

type GetCmd struct {
	CommonFlags
	Fields []string `arg:"" required:"" help:"json pointers to get."`
}

func (cmd *GetCmd) Run(cli *Context) error {
	f, err := newShadowFile(cmd.Path)
	if err != nil {
		return err
	}

	bs, err := lensed.Get(f.buf, cmd.Fields)
	if err != nil {
		return err
	}

	for _, b := range bs {
		fmt.Printf("%s\n", b)
	}
	return nil
}

type Setter struct {
	Field string
	Value string
}

func (s *Setter) UnmarshalText(in []byte) error {
	c := strings.SplitN(string(in), "=", 2)
	if len(c) != 2 {
		return fmt.Errorf("bad -v format %q, missing '='", in)
	}
	s.Field, s.Value = c[0], c[1]

	if strings.HasPrefix(s.Value, "@") {
		b, err := os.ReadFile(strings.TrimPrefix(s.Value, "@"))
		if err != nil {
			return err
		}
		s.Value = string(b)
	} else if strings.HasPrefix(s.Value, `\@`) {
		s.Value = strings.TrimPrefix(s.Value, `\`)
	}

	return nil
}

func version() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "(devel)"
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.UsageOnError(),
		kong.Vars{
			"version": version(),
		},
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)

	err := ctx.Run(&Context{CLI: &cli})
	ctx.FatalIfErrorf(err)
}
