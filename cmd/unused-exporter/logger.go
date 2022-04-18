package main

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

type logger struct {
	w io.Writer
}

func (l logger) Log(msg string, labels ...interface{}) {
	w := &bytes.Buffer{}

	fmt.Fprintf(w, "timestamp=%s msg=%q", time.Now().Format(time.RFC3339), msg)

	for i, j := 0, (len(labels)/2)*2; i < j; i += 2 {
		fmt.Fprintf(w, " %s=", labels[i])
		switch v := labels[i+1].(type) {
		case time.Time:
			fmt.Fprintf(w, "%q", v.Format(time.RFC3339))
		case error:
			fmt.Fprintf(w, "%q", v.Error())
		case string, fmt.Stringer:
			fmt.Fprintf(w, "%q", v)
		default:
			fmt.Fprint(w, v)
		}
	}

	if len(labels)%2 == 1 {
		fmt.Fprintf(w, " %s=MISSING", labels[len(labels)-1])
	}

	fmt.Fprintln(w)

	l.w.Write(w.Bytes())
}

func (l logger) Curry(msg string, labels ...interface{}) func(...interface{}) {
	return func(lbls ...interface{}) {
		labels = append(labels, lbls...)
		l.Log(msg, labels...)
	}
}
