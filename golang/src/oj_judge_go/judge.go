package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"golang.org/x/net/websocket"
)

// SubmittedData 接收/处理数据结构体
type SubmittedData struct {
	Username    string
	ProblemID   string
	Result      string // 处理返回
	Memory      string // 处理返回
	MemoryLimit string
	Time        string // 处理返回
	TimeLimit   string
	Lang        string
	Length      string // 处理返回
	Submitted   string
	Code        string
}

func main() {
	/* 接收代码数据 */
	http.Handle("/websocket", websocket.Handler(SubmittedDataHandler))
	err := http.ListenAndServe("127.0.0.1:8887", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// SubmittedDataHandler WebSocket处理函数
func SubmittedDataHandler(w *websocket.Conn) {
	var error error
	for {
		var submittedDataJSON string
		var submittedData SubmittedData

		fmt.Println("-----------------------------------")
		// 0.尝试接收消息
		error = websocket.Message.Receive(w, &submittedDataJSON)
		if error != nil {
			fmt.Println("0. 不能接收消息,error=", error)
			break
		}

		// 1.能接收消息
		fmt.Println("1. 接收消息：", submittedDataJSON)
		json.Unmarshal([]byte(submittedDataJSON), &submittedData)
		Username := submittedData.Username
		ProblemID := submittedData.ProblemID
		Result := ""
		MemoryLimit := submittedData.MemoryLimit
		Memory := ""
		TimeLimit := submittedData.TimeLimit
		Time := ""
		Lang := submittedData.Lang
		Length := strconv.Itoa(len(submittedData.Code)) + "B"
		Submitted := submittedData.Submitted
		SubmittedArray := strings.Split(submittedData.Submitted, " ")
		SubmittedTmp := SubmittedArray[0] + "-" + SubmittedArray[1]
		Code := submittedData.Code

		// 2.处理消息
		// 2.0判断语言
		fmt.Println("2. 处理判断：")
		if Lang != "C" && Lang != "C++" { // Language Error
			fmt.Println("Language Error")
			Result = "Language Error"
		} else {
			// 2.1写入源码
			writeResult := WriteSourceFile(ProblemID, Username, SubmittedTmp, Lang, Code)
			if writeResult != "success" { // Write File Error
				fmt.Println("[Judge System Error]: write source file failed")
				Result = "Judge System Error"
			} else {
				fmt.Println("[write result]: success")
				// 2.2判断正误
				judgeResult := ""
				judgeResult, Time, Memory = JudgeCmd(ProblemID, Username, SubmittedTmp, Lang, TimeLimit, MemoryLimit)

				fmt.Println("[judge result]: " + judgeResult)
				Result = judgeResult
			}
		}

		// 3.发送消息
		//连接的话 只能是 string类型的
		params := SubmittedData{Username, ProblemID, Result,
			Memory, MemoryLimit, Time, TimeLimit, Lang, Length, Submitted, Code}

		returnData, _ := json.Marshal(params)
		fmt.Println("3. 发给客户端：" + string(returnData))
		error = websocket.Message.Send(w, string(returnData))
		if error != nil {
			fmt.Println("-1. 不能发送消息，拜拜ヾ(•ω•`)o")
			// break
		}

		// 4.删除源码
		fmt.Println("4. 删除源码：")
		deleteResult := DeleteSourceFile(ProblemID, Username, SubmittedTmp, Lang)
		if deleteResult != "success" { // delete File Error
			fmt.Println("[Judge System Error]: delete source file failed")
		} else {
			fmt.Println("[delete result]: success")
		}
	}
}

// WriteSourceFile 1.1写入源码
func WriteSourceFile(ProblemID string, Username string, Submitted string, Lang string, Code string) (result string) {
	sourceFilename, _, _, _ := GetFilename(ProblemID, Username, Submitted, Lang)
	err := ioutil.WriteFile(sourceFilename, []byte(Code), 0777)
	if err != nil {
		// fmt.Println(err)
		return "fail"
	}
	// fmt.Println("write file successful")
	return "success"
}

// DeleteSourceFile 1.2删除源码
func DeleteSourceFile(ProblemID string, Username string, Submitted string, Lang string) (result string) {
	sourceFilename, exeFilename, outFilename, _ := GetFilename(ProblemID, Username, Submitted, Lang)

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

// JudgeCmd 2.判断正误
func JudgeCmd(ProblemID string, Username string, Submitted string, Lang string, TimeLimit string, MemoryLimit string) (result string, costTime string, costMemory string) {
	_, exeFilename, outFilename, filename := GetFilename(ProblemID, Username, Submitted, Lang)
	_, stdoutFilename := GetStdioFilename(ProblemID)

	limitTimeArray := strings.Split(TimeLimit, "MS")
	limitTimeString := limitTimeArray[0]
	limitTimeInt, _ := strconv.Atoi(limitTimeString)

	limitMemoryArray := strings.Split(MemoryLimit, "K")
	limitMemoryString := limitMemoryArray[0]
	limitMemoryInt, _ := strconv.Atoi(limitMemoryString)
	// fmt.Println("[limitMemoryInt]", limitMemoryInt)

	res := 0
	costMemory = ""
	start := time.Now()

	// CE RE TLE MLE
	res, costMemory = JudgeWithDocker(ProblemID, Lang, filename, limitTimeInt, limitMemoryInt)
	if res == 0 {
		existFlag, err := isFileExists(exeFilename)
		if err != nil {
			return "System Error", "", ""
		} else if existFlag == false {
			return "Compile Error", "", ""
		}
		existFlag, err = isFileExists(outFilename)
		if err != nil {
			return "System Error", "", ""
		} else if existFlag == false {
			return "Runtime Error", "", ""
		}
	} else if res == 1 {
		return "Memory Limit Exceeded", "", ""
	}

	duration := time.Since(start)
	durationMilliSeconds := duration.Nanoseconds() / 1e6
	costTimeString := strconv.FormatInt(durationMilliSeconds, 10)
	costTimeInt, _ := strconv.Atoi(costTimeString)
	costTimeInt = costTimeInt - 2800
	// fmt.Println("[costTimeString]:", costTimeString)
	if costTimeInt < 0 {
		costTimeInt = 0
		costTimeString = "0"
	} else {
		costTimeString = strconv.Itoa(costTimeInt)
	}
	costTime = costTimeString + "MS"

	if costTimeInt > limitTimeInt { // Time Limit Exceeded
		return "Time Limit Exceeded", "", ""
	}
	fmt.Println("[costTimeInt]:", costTimeInt, "...[limitTimeInt]:", limitTimeInt)
	// fmt.Println("[costTimeString]:", costTimeString)

	res = JudgeWAandOLE(outFilename, stdoutFilename)
	if res != 0 {
		if res == 1 {
			return "Wrong Answer", "", ""
		} else if res == 2 {
			return "Output Limit Exceeded", "", ""
		} else { // System Error: [can't get file size] or [Wrong Answer CMD Error]
			return "System Error", "", ""
		}
	}

	res = JudgePE(outFilename, stdoutFilename)
	if res != 0 {
		if res == 1 {
			return "Presentation Error", "", ""
		}
		// System Error: [Presentation CMD Error]
		return "System Error", "", ""
	}

	return "Accepted", costTime, costMemory
}

//JudgeWithDocker 2.1 编译   2.2 运行
func JudgeWithDocker(ProblemID string, Lang string, filename string, timeLimit int, memoryLimit int) (res int, costMemory string) {
	res = 0
	costMemory = ""
	imageName := "oj/myalpine:0.11"           //镜像名称
	containerName := "container_" + ProblemID //容器名称
	workDir := "/"                            //container工作目录
	containerDir := "/judgingFile"            //容器挂在目录
	hostDir := "/root/golang/src/oj_judge_go" //容器挂在到宿主机的目录
	judgeDir := containerDir + "/judging/"
	stdDir := containerDir + "/standard/"
	stdinDir := stdDir + "input/"
	stdinFilename := stdinDir + ProblemID + ".in"
	exeFilename := judgeDir + filename + ".e"
	outFilename := judgeDir + filename + ".out"
	sourceFilename := judgeDir + filename + ".c"
	cmdCR := "gcc -o " + exeFilename + " " + sourceFilename +
		" && " + exeFilename + " < " + stdinFilename + " > " + outFilename +
		" && sleep 5"
	if Lang == "C++" {
		sourceFilename = judgeDir + filename + ".cpp"
		cmdCR = "g++ -o " + exeFilename + " " + sourceFilename +
			" && " + exeFilename + " < " + stdinFilename + " > " + outFilename +
			" && sleep 5"
	}
	timeLimitDuration := time.Duration(5000) + time.Duration(timeLimit)*time.Millisecond //单位：ms
	cmd := []string{"sh", "-c", cmdCR}                                                   //运行的cmd命令，用于启动container中的程序
	// fmt.Println("[cmd]", cmdCR)

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	defer cli.Close()
	if err != nil {
		panic(err)
	}

	var tmp int64
	tmp = 0
	var resources container.Resources
	resources.Memory = 41943040 //int64(1024 * memoryLimit) // 40m = 41943040  单位：byte / B
	resources.MemorySwappiness = &tmp

	// 创建container
	cont, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      imageName, //镜像名称
		Cmd:        cmd,       //docker 容器中执行的命令
		WorkingDir: workDir,   //docker 容器中的工作目录
	}, &container.HostConfig{
		Resources: resources,
		Mounts: []mount.Mount{ //docker 容器目录挂在到宿主机目录
			{
				Type:   mount.TypeBind,
				Source: hostDir,
				Target: containerDir,
			},
		},
	}, nil, containerName)
	if err == nil {
		log.Printf("success create container:%s\n", cont.ID)
	} else {
		log.Println("failed to create container!", err)
	}

	// 运行container
	containerID := cont.ID
	err = cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err == nil {
		log.Printf("success start container:%s\n", containerID)
	} else {
		log.Printf("failed to start container:%s err:%v\n", containerID, err)
	}

	// 获取container stats
	statsTimeout := time.Second * 5
	statsCtx, _ := context.WithTimeout(context.Background(), statsTimeout)
	stats, err := cli.ContainerStats(statsCtx, containerID, false)
	if err != nil {
		log.Println(err)
	} else {
		statsJSON, err := ioutil.ReadAll(stats.Body)
		if err != nil {
			panic(err)
		}

		//json str 转map
		var statDataMap map[string]interface{}
		if err := json.Unmarshal([]byte(statsJSON), &statDataMap); err == nil {
			// fmt.Println(statDataMap)
			mapTmp := statDataMap["memory_stats"].(map[string]interface{})
			mapFloat64 := mapTmp["max_usage"].(float64)
			costMemoryFloat64 := mapFloat64 / 1024 / 2 // 单位：KB
			limitMemoryStr := strconv.Itoa(memoryLimit)
			limitMemoryFloat64, _ := strconv.ParseFloat(limitMemoryStr, 64)
			// fmt.Println("[0mapTmp]：", mapTmp, " mapFloat64:", mapFloat64)
			if costMemoryFloat64 == 0 {
				costMemory = "0K"
				fmt.Println("[1memory_max_usage]：", costMemory, " limit:", limitMemoryFloat64)
			} else if costMemoryFloat64 <= limitMemoryFloat64 { // 未超出
				costMemoryInt := int(costMemoryFloat64)
				costMemory = strconv.Itoa(costMemoryInt) + "K"
				fmt.Println("[2memory_max_usage]：", costMemoryInt, "KB , cost:", costMemory, " limit:", memoryLimit)
			} else {
				res = 1
				fmt.Println("[3memory_limit_error] limit:", limitMemoryFloat64)
			}
		}

		// log.Println(string(statsJSON))
		stats.Body.Close()
	}

	// 超过timeLimit(单位：ms) 暂停Container
	err = cli.ContainerStop(ctx, containerID, &timeLimitDuration)
	if err != nil {
		panic(err)
	}

	// 等待container运行结束
	_, err = cli.ContainerWait(ctx, containerID)
	if err != nil {
		panic(err)
	}

	// 删除container
	err = cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}

	return res, costMemory
}

// JudgeWAandOLE 2.3 调用命令行，比对输出文件(忽略空格造成的差异)
func JudgeWAandOLE(outFilename string, stdoutFilename string) (res int) {
	outFileSize := GetFileSize(outFilename)
	stoutputFileSize := GetFileSize(stdoutFilename)
	// fmt.Println("#### Output Size: ", stoutputFileSize, " ... ", outFileSize)
	if stoutputFileSize == 0 || outFileSize == 0 {
		return 3 // System Error: can't get file size
	}
	if stoutputFileSize*2 < outFileSize { // Judge Result: [Output Limit Exceeded]
		return 2
	}

	cmd := exec.Command("diff", "-b", outFilename, stdoutFilename)
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

// JudgePE 2.4 调用命令行，比对输出文件(并不忽略空格造成的差异)
func JudgePE(outFilename string, stdoutFilename string) (res int) {
	//比对是否正确
	cmd := exec.Command("diff", "-a", outFilename, stdoutFilename)
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

// GetFilename util1. 获取用户代码文件名
func GetFilename(ProblemID string, Username string, Submitted string, Lang string) (sourceFilename string, exeFilename string, outFilename string, filename string) {
	judgeDir := "judging/"
	filename = ProblemID + "_" + Username + "_" + Submitted
	sourceFilename = judgeDir + filename
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
	exeFilename = judgeDir + filename + ".e"
	outFilename = judgeDir + filename + ".out"
	return sourceFilename, exeFilename, outFilename, filename
}

// GetStdioFilename util2. 获取标准输入输出文件名
func GetStdioFilename(ProblemID string) (stdinFilename string, stdoutFilename string) {
	stdinFilename = "standard/input/" + ProblemID + ".in"
	stdoutFilename = "standard/output/" + ProblemID + ".out"
	return stdinFilename, stdoutFilename
}

// GetFileSize util3. 获取文件大小
func GetFileSize(filename string) (result int64) {
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}

//isFileExists util4. 判断文件是否存在
func isFileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
