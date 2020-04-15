package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/net/websocket"
)

type SubmittedData struct {
	Username  string
	ProblemID string
	Result    string // 处理返回
	Memory    string // 处理返回
	Time      string // 处理返回
	Lang      string
	Length    string // 处理返回
	Submitted string
	Code      string
}

func main() {
	/* 接收代码数据 */
	http.Handle("/websocket", websocket.Handler(SubmittedDataHandler))
	err := http.ListenAndServe("127.0.0.1:8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func SubmittedDataHandler(w *websocket.Conn) {
	var error error
	for {
		var submittedDataJson string
		var submittedData SubmittedData

		fmt.Println("-----------------------------------")
		// 0.尝试接收消息
		error = websocket.Message.Receive(w, &submittedDataJson)
		if error != nil {
			fmt.Println("0. 不能接收消息,error=", error)
			break
		}

		// 1.能接收消息
		fmt.Println("1. 接收消息：", submittedDataJson)
		json.Unmarshal([]byte(submittedDataJson), &submittedData)
		Username := submittedData.Username
		ProblemID := submittedData.ProblemID
		Result := ""
		Memory := ""
		Time := ""
		Lang := submittedData.Lang
		Length := strconv.Itoa(len(submittedData.Code)) + "B"
		Submitted := submittedData.Submitted
		Code := submittedData.Code

		// 2.处理消息
		// 2.0判断语言
		fmt.Println("2. 处理判断：")
		if Lang != "C" && Lang != "C++" { // Language Error
			fmt.Println("Language Error")
			Result = "Language Error"
		} else {
			// 2.1写入源码
			writeResult := WriteSourceFile(ProblemID, Username, Submitted, Lang, Code)
			if writeResult != "success" { // Write File Error
				fmt.Println("[Judge System Error]: write source file failed")
				Result = "Judge System Error"
			} else {
				fmt.Println("[write result]: success")
				// 2.2判断正误
				start := time.Now()
				judgeResult := JudgCmd(ProblemID, Username, Submitted, Lang)
				cost := time.Since(start)
				fmt.Println("!!!!!!!!!!!!!cost=[%s]", cost)

				fmt.Println("[judge result]: " + judgeResult)
				Result = judgeResult
			}
		}

		// 3.发送消息
		//连接的话 只能是 string类型的
		params := SubmittedData{Username, ProblemID, Result,
			Memory, Time, Lang, Length, Submitted, Code}

		returnData, _ := json.Marshal(params)
		fmt.Println("3. 发给客户端：" + string(returnData))
		error = websocket.Message.Send(w, string(returnData))
		if error != nil {
			fmt.Println("-1. 不能发送消息，拜拜ヾ(•ω•`)o")
			// break
		}

		// 4.删除源码
		fmt.Println("4. 删除源码：")
		deleteResult := DeleteSourceFile(ProblemID, Username, Submitted, Lang)
		if deleteResult != "success" { // delete File Error
			fmt.Println("[Judge System Error]: delete source file failed")
		} else {
			fmt.Println("[delete result]: success")
		}
	}
}

/* 1.1写入源码 */
func WriteSourceFile(ProblemID string, Username string, Submitted string, Lang string, Code string) (result string) {
	sourceFilename, _, _ := GetFilename(ProblemID, Username, Submitted, Lang)
	err := ioutil.WriteFile(sourceFilename, []byte(Code), 0777)
	if err != nil {
		// fmt.Println(err)
		return "fail"
	} else {
		// fmt.Println("write file successful")
		return "success"
	}
}

/* 1.2删除源码 */
func DeleteSourceFile(ProblemID string, Username string, Submitted string, Lang string) (result string) {
	sourceFilename, exeFilename, outFilename := GetFilename(ProblemID, Username, Submitted, Lang)

	cmd := exec.Command("")
	cmd = exec.Command("rm", "-f", sourceFilename, exeFilename, outFilename)

	//运行命令
	err := cmd.Start()
	if err != nil { // CMD错误 Delete CMD Error
		return "fail"
	}
	//等待完成
	err = cmd.Wait()
	if err != nil { // 删除错误 Delete Error
		return "fail"
	}
	return "success"
}

/* 2.判断正误 */
func JudgCmd(ProblemID string, Username string, Submitted string, Lang string) (result string) {
	sourceFilename, exeFilename, outFilename := GetFilename(ProblemID, Username, Submitted, Lang)
	stinputFilename, stoutputFilename := GetStioFilename(ProblemID)

	res := JudgeCE(sourceFilename, exeFilename, Lang)
	if res != 0 {
		if res == 1 {
			return "Compile Error"
		} else { // System Error: [language error] or [Compile CMD Error]
			return "System Error"
		}
	}
	res = JudgeRE(exeFilename, outFilename, stinputFilename)
	if res != 0 {
		if res == 1 {
			return "Runtime Error"
		} else { // System Error: [Runtime CMD Error]
			return "System Error"
		}
	}

	res = JudgeWAandOLE(outFilename, stoutputFilename)
	if res != 0 {
		if res == 1 {
			return "Wrong Answer"
		} else if res == 2 {
			return "Output Limit Exceeded"
		} else { // System Error: [can't get file size] or [Wrong Answer CMD Error]
			return "System Error"
		}
	}

	res = JudgePE(outFilename, stoutputFilename)
	if res != 0 {
		if res == 1 {
			return "Presentation Error"
		} else { // System Error: [Presentation CMD Error]
			return "System Error"
		}
	}

	return "Accepted"
}

/* 2.1 调用命令行，编译代码文件 */
func JudgeCE(sourceFilename string, exeFilename string, Lang string) (res int) {
	cmd := exec.Command("")
	if Lang == "C" {
		cmd = exec.Command("gcc", "-o", exeFilename, sourceFilename)
	} else if Lang == "C++" {
		cmd = exec.Command("g++", "-o", exeFilename, sourceFilename)
		// } else if Lang == "Python3" {
		// 	return 0
	} else { // System Error: [language error]
		return 2
	}
	//运行命令
	err := cmd.Start()
	if err != nil { // System Error: [Compile CMD Error]
		return 2
	}
	//等待完成
	err = cmd.Wait()
	if err != nil { // Judge Result: [Compile Error] 编译错误
		return 1
	}
	return 0 // Continue
}

/* 2.2 调用命令行，运行代码文件 */
func JudgeRE(exeFilename string, outFilename string, stinputFilename string) (res int) {
	fin, _ := os.OpenFile("./"+stinputFilename, os.O_RDONLY, 0755)
	fout, _ := os.OpenFile("./"+outFilename, os.O_WRONLY|os.O_CREATE, 0755)

	cmd := exec.Command("./" + exeFilename)
	cmd.Stdin = fin
	cmd.Stdout = fout

	//运行命令
	err := cmd.Start()
	if err != nil { // or System Error: [Runtime CMD Error]
		return 2
	}
	//等待完成
	err = cmd.Wait()
	if err != nil { // Judge Result: [Runtime Error] 运行错误
		return 1
	}
	return 0 // Continue
}

/* 2.3 调用命令行，比对输出文件(忽略空格造成的差异) */
func JudgeWAandOLE(outFilename string, stoutputFilename string) (res int) {
	outFileSize := GetFileSize(outFilename)
	stoutputFileSize := GetFileSize(stoutputFilename)
	fmt.Println("#### Output Size: ", stoutputFileSize, " ... ", outFileSize)
	if stoutputFileSize == 0 || outFileSize == 0 {
		return 3 // System Error: can't get file size
	}
	if stoutputFileSize*2 < outFileSize { // Judge Result: [Output Limit Exceeded]
		return 2
	}

	cmd := exec.Command("diff", "-b", outFilename, stoutputFilename)
	//运行命令
	err := cmd.Start()
	if err != nil { // System Error: [Wrong Answer CMD Error]
		return 3
	}
	//等待完成
	err = cmd.Wait()
	if err != nil { // Judge Result: [Wrong Answer]
		return 1
	}
	return 0 // Continue
}

/* 2.4 调用命令行，比对输出文件(并不忽略空格造成的差异) */
func JudgePE(outFilename string, stoutputFilename string) (res int) {
	//比对是否正确
	cmd := exec.Command("diff", "-a", outFilename, stoutputFilename)
	//运行命令
	err := cmd.Start()
	if err != nil { // System Error: [Presentation CMD Error]
		return 2
	}
	//等待完成
	err = cmd.Wait()
	if err != nil { // Judge Result: Presentation Error
		return 1
	}
	return 0 // Judge Result: Accept
}

/* util1. 获取用户代码文件名 */
func GetFilename(ProblemID string, Username string, Submitted string, Lang string) (sourceFilename string, exeFilename string, outFilename string) {
	filename := "judging/" + ProblemID + "_" + Username + "_" + Submitted
	sourceFilename = filename
	switch Lang {
	case `C`:
		sourceFilename += ".c"
	case `C++`:
		sourceFilename += ".cpp"
		// case `Python3`:
		// 	sourceFilename += ".py"
		// default:
		// 	sourceFilename += ".c"
	}
	exeFilename = filename + ".e"
	outFilename = filename + ".out"
	return sourceFilename, exeFilename, outFilename
}

/* util2. 获取标准输入输出文件名 */
func GetStioFilename(ProblemID string) (stinputFilename string, stoutputFilename string) {
	stinputFilename = "standard/input/" + ProblemID + ".in"
	stoutputFilename = "standard/output/" + ProblemID + ".out"
	return stinputFilename, stoutputFilename
}

/* util3. 获取文件大小 */
func GetFileSize(filename string) (result int64) {
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}
