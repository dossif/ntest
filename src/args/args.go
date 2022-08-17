package args

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/go-playground/validator/v10"
	"os"
	"strconv"
)

type Arguments struct {
	IcmpPing IcmpPing
	TcpTest  TcpTest
}

type IcmpPing struct {
	Command *argparse.Command
	Flags   struct {
		Host     *string
		Timeout  *int
		Dns      *string
		Bind     *string
		Interval *int
		Warn     *int
	}
}

type TcpTest struct {
	Command *argparse.Command
	Flags   struct {
		Host     *string
		Port     *int
		Timeout  *int
		Interval *int
		Dns      *string
		Bind     *string
	}
}

func NewArgs(appName string, appVersion string) Arguments {
	parser := argparse.NewParser(appName, fmt.Sprintf("%v v%v network testing tool", appName, appVersion))
	var args Arguments
	// common flags
	hostFlagOpt := &argparse.Options{
		Required: true,
		Help:     "target host",
		Default:  "127.0.0.1",
		Validate: func(args []string) error {
			return validateArg(args[0], "string", "hostname|ip4_addr|ip6_addr")
		},
	}
	portFlagOpt := &argparse.Options{
		Required: false,
		Help:     "target port",
		Default:  80,
	}
	timeoutFlagOpt := &argparse.Options{
		Required: false,
		Help:     "request timeout in ms",
		Default:  3000,
		Validate: func(args []string) error { return validateArg(args[0], "int", "gte=0") },
	}
	dnsFlagOpt := &argparse.Options{
		Required: false,
		Help:     "dns server",
		Validate: func(args []string) error { return validateArg(args[0], "string", "hostname|ip4_addr|ip6_addr") },
	}
	bindFlagOpt := &argparse.Options{
		Required: false,
		Help:     "bind address",
		Default:  "0.0.0.0",
		Validate: func(args []string) error { return validateArg(args[0], "string", "hostname|ip4_addr|ip6_addr") },
	}
	intervalFlagOpt := &argparse.Options{
		Required: false,
		Help:     "interval in ms",
		Default:  1000,
		Validate: func(args []string) error { return validateArg(args[0], "int", "gt=0") },
	}
	warnFlagOpt := &argparse.Options{
		Required: false,
		Help:     "warn threshold in ms",
		Default:  100,
		Validate: func(args []string) error { return validateArg(args[0], "int", "gt=0") },
	}
	// icmp ping
	parser.DisableHelp()
	args.IcmpPing.Command = parser.NewCommand("ping", "host icmp ping")
	args.IcmpPing.Flags.Host = args.IcmpPing.Command.String("h", "host", hostFlagOpt)
	args.IcmpPing.Flags.Timeout = args.IcmpPing.Command.Int("t", "timeout", timeoutFlagOpt)
	args.IcmpPing.Flags.Dns = args.IcmpPing.Command.String("d", "dns", dnsFlagOpt)
	args.IcmpPing.Flags.Bind = args.IcmpPing.Command.String("b", "bind", bindFlagOpt)
	args.IcmpPing.Flags.Interval = args.IcmpPing.Command.Int("i", "interval", intervalFlagOpt)
	args.IcmpPing.Flags.Warn = args.IcmpPing.Command.Int("w", "warn", warnFlagOpt)
	// tcp test
	args.TcpTest.Command = parser.NewCommand("tcp", "host tcp-port test")
	args.TcpTest.Flags.Host = args.TcpTest.Command.String("h", "host", hostFlagOpt)
	args.TcpTest.Flags.Port = args.TcpTest.Command.Int("p", "port", portFlagOpt)
	args.TcpTest.Flags.Timeout = args.TcpTest.Command.Int("t", "timeout", timeoutFlagOpt)
	args.TcpTest.Flags.Interval = args.TcpTest.Command.Int("i", "interval", intervalFlagOpt)
	args.TcpTest.Flags.Dns = args.TcpTest.Command.String("d", "dns", dnsFlagOpt)
	args.TcpTest.Flags.Bind = args.TcpTest.Command.String("b", "bind", bindFlagOpt)
	parser.SetHelp("", "help")
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}
	return args
}

func validateArg(name string, vType string, tag string) (err error) {
	var val interface{}
	switch vType {
	case "string":
		val = name
	case "int":
		val, err = strconv.Atoi(name)
		if err != nil {
			return fmt.Errorf("failed to convert %v to int", name)
		}
	}
	v := validator.New()
	err = v.Var(val, tag)
	if err != nil {
		return fmt.Errorf("failed to validate value \"%v\", must be %v", name, tag)
	}
	return err
}
