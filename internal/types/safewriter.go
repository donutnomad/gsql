package types

import "github.com/donutnomad/gsql/clause"

type SafeWriter struct {
	Builder clause.Builder
}

func (w *SafeWriter) WriteByte(b byte) error {
	return w.Builder.WriteByte(b)
}
func (w *SafeWriter) WriteString(b string) {
	_, _ = w.Builder.WriteString(b)
}
func (w *SafeWriter) WriteQuoted(f any) {
	w.Builder.WriteQuoted(f)
}
func (w *SafeWriter) AddVar(writer *SafeWriter, args ...any) {
	w.Builder.AddVar(writer.Builder, args...)
}
