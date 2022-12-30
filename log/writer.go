package log

type LogWriter interface {
	Levels() []Level
	LogWrite(b []byte) error
	Close()
}
type LevelWriters map[Level][]LogWriter

func (w LevelWriters) Add(logWriter LogWriter) {
	for _, level := range logWriter.Levels() {
		w[level] = append(w[level], logWriter)
	}
}
func (w LevelWriters) Fire(level Level, b []byte) error {
	for _, hook := range w[level] {
		if err := hook.LogWrite(b); err != nil {
			return err
		}
	}
	return nil
}
