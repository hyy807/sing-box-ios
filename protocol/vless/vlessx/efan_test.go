package vlessx
import("bytes";"testing"; M "github.com/sagernet/sing/common/metadata")
func TestEFANTCPFrame(t *testing.T){
 var b bytes.Buffer
 req:=Request{X365:true,Command:1,Destination:M.ParseSocksaddr("example.com:443")}
 for i:=range req.UUID {req.UUID[i]=byte(i)}
 if err:=WriteRequest(&b,req,[]byte("GET"));err!=nil{t.Fatal(err)}
 want:=append([]byte{'X','3','6','5',1,1},req.UUID[:]...)
 want=append(want,1,187,2,11)
 want=append(want,[]byte("example.comGET")...)
 if !bytes.Equal(b.Bytes(),want){t.Fatalf("wire mismatch: got %x want %x",b.Bytes(),want)}
}
func TestEFANResponse(t *testing.T){
 for _,tc:=range []struct{p []byte;ok bool}{{[]byte{'X','3','6','5',0},true},{[]byte{'X','3','6','5',1},false},{[]byte("wrong"),false},{[]byte("X"),false}} {
 if err:=ReadResponse(bytes.NewReader(tc.p),true);(err==nil)!=tc.ok{t.Fatalf("response %x: %v",tc.p,err)}
 }
}
