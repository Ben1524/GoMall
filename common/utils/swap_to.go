package utils
import "github.com/bytedance/sonic"


//通过json tag 进行结构体赋值
func SwapTo(request, category interface{}) (err error)  {
	dataByte, err := sonic.Marshal(request)
	if err != nil {
		return
	}
	err = sonic.Unmarshal(dataByte, category)
	return
}
