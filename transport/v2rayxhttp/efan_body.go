package v2rayxhttp

import (
 "bufio"
 "io"
 "unicode/utf8"
)

type latin1Writer struct { w io.Writer }
func (l *latin1Writer) Write(p []byte) (int,error) {
 encoded:=make([]byte,0,len(p)*2)
 for _, b:=range p { encoded=utf8.AppendRune(encoded,rune(b)) }
 for len(encoded)>0 { n,err:=l.w.Write(encoded); if err!=nil{return 0,err}; if n==0{return 0,io.ErrShortWrite}; encoded=encoded[n:] }
 return len(p),nil
}
type latin1Reader struct { r *bufio.Reader; closer io.Closer }
func (l *latin1Reader) Read(p []byte)(int,error) {
 if len(p)==0{return 0,nil}
 ch,_,err:=l.r.ReadRune(); if err!=nil{return 0,err}
 if ch>255{return 0,io.ErrUnexpectedEOF}
 p[0]=byte(ch); return 1,nil
}

func (l *latin1Writer) Close() error { if c,ok:=l.w.(io.Closer);ok{return c.Close()};return nil }
func (l *latin1Reader) Close() error { if l.closer!=nil{return l.closer.Close()};return nil }
