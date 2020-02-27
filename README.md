cwcmd
=====


[![Go Report Card][goreport-img]][goreport-url]
[![GoDoc][godoc-img]][godoc-url]


cwcmd is a simple wrapper around go-cmd/cmd package. Its main purpose is
to simplify its usage (if this is possible) and add a generic purpose
hook.

The typical usage of this hook is to perform all logging operations and
stream the command output to the terminal.

# Example

```go
func main() {
	c := cwcmd.New(&cwcmd.Options{
		Buffered:  false,
		Streaming: true,
	}, "ls", "-l", "/")
 
	c.AddHook(PrintLines)
	if err := c.Start(); err != nil {
		panic(err)
	}
 
	i, err := c.Wait()
	fmt.Printf("Done: %d\n", i)
	if err != nil {
		panic(err)
	}
}
  
func PrintLines(h *cwcmd.Hook) {
	fmt.Printf("Starting command: %s %s\n", h.Cmd.Name, h.Cmd.Args)
	defer close(h.Done)
	for h.Cmd.Stdout != nil || h.Cmd.Stderr != nil {
		select {
		case line, open := <-h.Cmd.Stdout:
			if !open {
				h.Cmd.Stdout = nil
				continue
			}
			fmt.Println(line)
		case line, open := <-h.Cmd.Stderr:
			if !open {
				h.Cmd.Stderr = nil
				continue
			}
			fmt.Fprintln(os.Stderr, line)
		}
	}
	fmt.Printf("Exit: %d\n", h.Cmd.Status().Exit)
}
```


# License

Copyright © 2020 Sébastien Gross <seb•ɑƬ•chezwam•ɖɵʈ•org> 

This program is free software. It comes without any warranty, to the extent
permitted by applicable law. You can redistribute it and/or modify it under
the terms of the Do What The Fuck You Want To Public License, Version 2, as
published by Sam Hocevar. See http://sam.zoy.org/wtfpl/COPYING for more
details.


[goreport-img]: https://goreportcard.com/badge/github.com/renard/cwcmd
[goreport-url]: https://goreportcard.com/report/github.com/renard/cwcmd
[godoc-img]: https://godoc.org/github.com/renard/cwcmd?status.svg
[godoc-url]: https://godoc.org/github.com/renard/cwcmd
