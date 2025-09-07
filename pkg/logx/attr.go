package logx

import "log/slog"

func StateName(name string) slog.Attr {
	return slog.String("state.name", name)
}

func CommandName(name string) slog.Attr {
	return slog.String("command.name", name)
}

func Err(err error) slog.Attr {
	return slog.Any("err", err)
}
