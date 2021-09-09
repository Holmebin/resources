package main


//#cgo LDFLAGS:-L./ -lSSCardDriver -Wl,-rpath .
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include "SSCardDriver.h"
import "C"
import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"golang.org/x/net/websocket"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"unicode/utf16"
	"unsafe"
)
var url ="http://10.84.94.160:7023/lskzx/SiInterfaceAction.do?"
//var url  ="http://10.87.0.72:9006/WebServiceTz/services/callBusiness"
var server = websocket.Server{Handshake: nil}
func main() {
	//action2001, _ := SiInterfaceAction2001("")
	//fmt.Println( siInterfaceAction("") )
	//fmt.Println(action2001.ParaValue)
	//CardLSDll()
	//fmt.Println(CallBusiness("331100|00904A3001866033110000048A|03|331100D156000005500509728DB46EF9|456FE02E253F817D|A39FC0BF44CAEA80|4F5FB5932F10694D|2CFE760428888D8E|"))
	//server.Handler=Echo
	//http.Handle("/echo", server)
	server.Handler=ReadCard
	http.Handle("/readCard", server)
	if err := http.ListenAndServe(":8089", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}


}

func ReadCard(ws *websocket.Conn) {
	var err error
	for {
		var reply string
		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive")
			break
		}

		msg,err := SiInterfaceAction2001(reply)
		if err != nil {
			msg.ParaValue=err.Error()
		}
		fmt.Println(reply+"^^^"+msg.ParaValue)
		if err = websocket.Message.Send(ws, msg.ParaValue); err != nil {
			fmt.Println("Can't send")
			break
		}
	}
}
func Cpinterface(arg0 string,arg1 string) string {
	arg0= If(arg0 == "","2001",arg0).(string)
	arg1= If(arg1 == "","2|1001|1|02|B8C37E33DEFDE51CF91E1E03E51657DA|1001|1111111111|331102001|",arg1).(string)
	var builder strings.Builder
	builder.WriteString("method=siInterface&busType=1&eap_password=0000&eap_username=interface&operType=1&inputData=")
	builder.WriteString(arg0)
	builder.WriteString("^1001^0000^^^^")
	builder.WriteString(arg1)
	dll, err :=CardOs()
	if err != nil {
		return "调用os失败"
	}
	builder.WriteString(dll)
	builder.WriteString("||^")
	action, err := SiInterfaceAction(builder.String())
	if err != nil {
		action.ParaValue=err.Error()
	}
	return action.ParaValue
}
func SiInterfaceAction2001(arg0 string)(Parameter,error) {
	arg0= If(arg0 == "","2|1001|1|02|B8C37E33DEFDE51CF91E1E03E51657DA|1001|1111111111|331102001|",arg0).(string)
	var builder strings.Builder
	builder.WriteString("method=siInterface&busType=1&eap_password=0000&eap_username=interface&operType=1&inputData=2001^1001^0000^^^^")
	builder.WriteString(arg0)
	fmt.Println("CardOs")
	dll, err := CardOs()
	if err != nil {
		return Parameter{ParaValue: "调用os失败"},err
	}
	builder.WriteString(dll)
	builder.WriteString("||^")
	fmt.Println("SiInterfaceAction2001")
	return SiInterfaceAction(builder.String())
}
func SiInterfaceAction8004(arg0 string)(Parameter,error) {
	var builder strings.Builder
	builder.WriteString("method=siInterface&busiType=1&cardInfo=&inputData=8004^000000000000^^^^0000^")
	builder.WriteString(arg0)
	builder.WriteString("^1^")
	return SiInterfaceAction(builder.String())
}
func SiInterfaceAction(reqBody string)(Parameter,error) {
	v := Parameter{}
	toGbk,_ := Utf8ToGbk([]byte(reqBody))
	res, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewReader(toGbk))
	if nil != err {
		return v,&HttpErr{-1,"http post err"}
	}

	fmt.Println("SiInterfaceAction")
	defer res.Body.Close()
	fmt.Println(res.StatusCode)
	//if http.StatusOK != res.StatusCode {
	//	return v,&HttpErr{-1,"status err"}
	//}
	data, err := ioutil.ReadAll(res.Body)
	if nil != err {
		return v,err
	}
	gbk, _ := GbkToUtf8(data)
	fmt.Println(ByteString(gbk))
	re := ReponseEnvelope{}
	xml.Unmarshal(gbk, &re)
	fmt.Println(re)
	parameters := re.ReponseEnvelopeBody.Parameters.Parameter
	for i := range parameters {
		if parameters[i].ParaName=="outputData" {
			//^^98ABD5279E9CD3306183AFC81BE3C72E|40B69F2C478028DC1594A8CC187E1FE9|^
			//println(parameters[i].ParaValue)
			return parameters[i],nil
		}
	}
	return v,&HttpErr{-1,"notfound"}
}
func CardOs()(string,error){
	pOutInfo:= C.CString(string(make([]byte, 1024)))
	i := C.iReadCardBas(1,pOutInfo)
	defer C.free(unsafe.Pointer(pOutInfo))
	fmt.Println("i load_iReadCardBas:",i)
	b := C.GoString(pOutInfo)
	utf8, _ := GbkToUtf8([]byte(b))
	fmt.Println(ByteString(utf8)+"111")
	if  i== -2201 || i==-2202 || i==-2203 || i==-27272 || i==-24 {
		pOutInfo2 := C.CString(string(make([]byte, 1024)))
		i = C.iReadCardBas_HSM_Step1(1,pOutInfo2)
		defer C.free(unsafe.Pointer(pOutInfo2))
		fmt.Println("i load_iReadCardBas_HSM_Step1:",i)
		b =C.GoString(pOutInfo2)
		utf8, _ := GbkToUtf8([]byte(b))
		fmt.Println(ByteString(utf8)+"111")
		if i != 0 {
			return ByteString(utf8),&HttpErr{Code:-1}
		}
		action8004, err := SiInterfaceAction8004(ByteString(utf8))
		fmt.Println("err",err)
		if err == nil {
			var result string
			result=string([]rune(action8004.ParaValue)[2:68])
			e := C.CString(result)
			pOutInfo = C.CString(string(make([]byte, 1024)))
			i = C.iReadCardBas_HSM_Step2(e,pOutInfo)
			fmt.Println("i load_iReadCardBas_HSM_Step2:",i)
			b = C.GoString(pOutInfo)
			utf8, _ := GbkToUtf8([]byte(b))
			fmt.Println(ByteString(utf8)+"111")
			if i != 0 {
				return ByteString(utf8),&HttpErr{Code:-1}
			}
			//b, _ = GbkToUtf8(b)
			return  ByteString(utf8),nil
		}
	}
	if i == 0{
		//utf8, err := GbkToUtf8(b)
		return ByteString(utf8),nil
	}
	return "读卡错误",&HttpErr{Code:-1}
}
/*func CardDll()(string,error){
	h, err := syscall.LoadDLL("SSCardDriver.dll")
	if err != nil {
		return "Can't LoadLibrary",err
	}
	proc, _ := h.FindProc("iReadCardBas")
	b := make([]byte, 4028)
	c1, _, err := proc.Call( 1,  uintptr(unsafe.Pointer(&b[0])))
	i := int32(c1)
	if  i== -2201 || i==-2202 || i==-2203 || i==-27272 || i==-24 {
		ireadcardbasHsmStep1, _ := h.FindProc("iReadCardBas_HSM_Step1")
		b = make([]byte, 4028)
		call, _, err := ireadcardbasHsmStep1.Call(1, uintptr(unsafe.Pointer(&b[0])))
		if int32(call) != 0 {
			return  string(int32(call)),err
		}
		b, _ = GbkToUtf8(b)
		action8004, err := SiInterfaceAction8004(ByteString(b))
		//envelope,err := CallbusinessFzyw9013(ByteString(b))
 		//result , _ := json.Marshal(v)
		if err == nil {
			var result string
			result=string([]rune(action8004.ParaValue)[2:68])
 			//result := root.RESULT1+"|"+root.RESULT2+"|"
			b = make([]byte, 4028)
			ireadcardbasHsmStep2, _ := h.FindProc("iReadCardBas_HSM_Step2")
			r1, _, err := ireadcardbasHsmStep2.Call(BytePtr([]byte(result)), BytePtr(b))
			if int32(r1) != 0 {
				return  string(int32(r1)),err
			}
			b, _ = GbkToUtf8(b)
			return  ByteString(b),nil
		}
	}
	if i == 0{
		utf8, err := GbkToUtf8(b)
		return ByteString(utf8),err
	}
	return "读卡错误",err
}*/
//reqBody = "method=siInterfaceAction&busiType=1&cardInfo=&inputData=8004^000000000000^^^^0000^331100|00904A3001866033110000048A|03|331100D156000005500509728DB46EF9|6CA9A3FD542C34C0|5550270144AD1FF3|E631C5F12A962D37|EB7F91AD7E4F03FF|^1^"


/*	reqBody := `<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xmlns:xsd="http://www.w3.org/2001/XMLSchema"
  xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetSpeech xmlns="http://xmlme.com/WebServices">
      <Request>string</Request>
    </GetSpeech>
  </soap:Body>
</soap:Envelope>` */
func CallbusinessFzyw9013(arg0 string)(Envelope,error ) {
	in0s := strings.Split(arg0, "|")
	//"331100|00904A3001866033110000048A|03|331100D156000005500509728DB46EF9|FF886A4907437B13|0D321F7B101B1281|586A95874C490C0D|D35A6572B44836DF|"

	var builder strings.Builder
	builder.WriteString("<soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:top=\"http://topcheer.com\"><soapenv:Header/><soapenv:Body><top:callBusiness_fzyw><top:in0><![CDATA[<ROOT><AAB301>")
	builder.WriteString( in0s[0] )
	builder.WriteString( "</AAB301><AAZ507>" )
	builder.WriteString( in0s[1] )
	builder.WriteString( "</AAZ507><AFLAG>" )
	builder.WriteString( in0s[2] )
	builder.WriteString( "</AFLAG><AAZ501>" )
	builder.WriteString( in0s[3] )
	builder.WriteString( "</AAZ501><SJS1>" )
	builder.WriteString( in0s[4] )
	builder.WriteString( "</SJS1><SJS2>" )
	builder.WriteString( in0s[5] )
	builder.WriteString( "</SJS2><SJS3>" )
	builder.WriteString( in0s[6] )
	builder.WriteString( "</SJS3><SJS4>" )
	builder.WriteString( in0s[7] )
	builder.WriteString( "</SJS4><MSGNO>9013</MSGNO></ROOT>]]></top:in0></top:callBusiness_fzyw></soapenv:Body></soapenv:Envelope>" )
	s := builder.String()
	return CallBusiness(s)
}


func CallBusiness(arg0 string)(Envelope,error ) {
	url  = "http://10.87.0.72:9006/WebServiceTz/services/CallBusiness"
	res, err := http.Post(url, "text/xml;charset=UTF-8", strings.NewReader(arg0))
	v := Envelope{}
	if nil != err {
		return v, err
	}
	defer res.Body.Close()
	// return status
	if http.StatusOK != res.StatusCode {
		return v, &HttpErr{-1, "request fail"}
	}
	data, err := ioutil.ReadAll(res.Body)
	if nil != err {
		return v, err
	}
	content := strings.Replace(string(data), "&lt;", "<", -1)
	err = xml.Unmarshal([]byte(content), &v)
	return v,nil
}

/*func CardLSDll()(s string){
	h, err := syscall.LoadDLL("LSCP.dll")
	if err != nil {
		fmt.Println("Can't LoadLibrary")
		return ""
	}
	proc, err := h.FindProc("cpinterface")
	if err != nil {
		fmt.Println("Can't GetProcAddress")
		return ""
	}
 	a := make([]byte, 255)
 	b := make([]byte, 255)
 	r, _, _ := proc.Call(StrPtr("2001"), StrPtr("1|1001|1|02|B8C37E33DEFDE51CF91E1E03E51657DA|1001|1111111111|331102001|"), uintptr(unsafe.Pointer(&a[0])),  uintptr(unsafe.Pointer(&b[0])))
	fmt.Println("rrr",r)
	a, err = GbkToUtf8(a)
	b, err = GbkToUtf8(b)
	fmt.Println(ByteString(a))
	fmt.Println(ByteString(b))
	return ByteString(a)
}*/
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func IntPtr(n int) uintptr {
	return uintptr(n)
}
func BytePtr(b []byte) uintptr {
	return uintptr(unsafe.Pointer(&b[0]))
}


func StrPtr(s string) uintptr {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return uintptr(0)
		}
	}
	encode := utf16.Encode([]rune(s + "\x00"))
	return uintptr(unsafe.Pointer(&encode[0]))
}

func ByteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
type Parameter struct {
	XMLName   xml.Name `xml:"parameter"`
	ParaName  string `xml:"paraName,attr"`
	ParaValue string `xml:"paraValue,attr"`
}

type  ErrorMessage struct{
	XMLName   xml.Name `xml:"errorMessage"`
	BriefMessage string `xml:"briefMessage,attr"`
	DetailMessage string `xml:"detailMessage,attr"`
}

type Parameters struct {
	XMLName   xml.Name `xml:"parameters"`
	Parameter []Parameter `xml:"parameter"`
}

type ReponseEnvelopeBody struct {
	XMLName    xml.Name `xml:"body"`
	DataStores string `xml:"dataStores"`
	Parameters Parameters `xml:"parameters"`
}
type ReponseEnvelopeHeader struct {
	XMLName    xml.Name `xml:"header"`
	AppCode    int `xml:"appCode"`
	ErrorMessage ErrorMessage `xml:"errorMessage"`
}
type ReponseEnvelope struct {
	XMLName xml.Name `xml:"reponseEnvelope"`
	ReponseEnvelopeBody ReponseEnvelopeBody `xml:"body"`
	ReponseEnvelopeHeader ReponseEnvelopeHeader `xml:"header"`
}

type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body Body `xml:"Body"`
}

type Body struct {
	XMLName  xml.Name  `xml:"Body"`
	CallBusiness_fzywResponse CallBusiness_fzywResponse `xml:"callBusiness_fzywResponse"`
}

type CallBusiness_fzywResponse struct {
	XMLName xml.Name `xml:"callBusiness_fzywResponse"`
	Out Out `xml:"out"`
}
type Out struct {
	XMLName xml.Name `xml:"out"`
	Root Root `xml:"ROOT"`
}
type Root struct {
	XMLName xml.Name `xml:"ROOT"`
	RETCODE string `xml:"RETCODE"`
	RETTEXT string `xml:"RETTEXT"`
	RESULT string `xml:"RESULT"`
	RESULT1 string `xml:"RESULT1"`
	RESULT2 string `xml:"RESULT2"`
}
type HttpErr struct {
	Code int
	Msg   string
}

func (e *HttpErr) Error() string {
	err, _ := json.Marshal(e)
	return string(err)
}