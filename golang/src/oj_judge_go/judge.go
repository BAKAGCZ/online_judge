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
	"strings"
	"golang.org/x/net/websocket"
	/*
	// "context"
	"io"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
    "github.com/docker/go-connections/nat"
    "golang.org/x/net/context"
	// "github.com/docker/docker/pkg/stdcopy"
	*/
)	

// cmd := exec.Command("docker", "run", "-it", "-v", "~/golang/src/oj_judge_go:/judgeFile oj/myalpine")

/*
const (
    imageName     string   = "oj/myalpine"				  //镜像名称
    containerName string   = "myalpine"			//容器名称
    indexName     string   = "/" + containerName	  //容器索引名称，用于检查该容器是否存在是使用
    cmd           string   = "./judging/4_root_2020-05-09 22:59:00.e"         //运行的cmd命令，用于启动container中的程序
    workDir       string   = "/judgeFile"		 			   //container工作目录
    openPort      nat.Port = "7070"							  //container开放端口
    hostPort      string   = "7070"								//container映射到宿主机的端口
    containerDir  string   = "/judgeFile"					  //容器挂在目录
    hostDir       string   = "~/golang/src/oj_judge_go" //容器挂在到宿主机的目录
    n             int      = 5                         		           //每5s检查一个容器是否在运行
 
)
*/

type SubmittedData struct {
	Username  string
	ProblemID string
	Result    string // 处理返回
	Memory    string // 处理返回
	MemoryLimit string
	Time      string // 处理返回
	TimeLimit string
	Lang      string
	Length    string // 处理返回
	Submitted string
	Code      string
}

func main() {
	/* 接收代码数据 */
	http.Handle("/websocket", websocket.Handler(SubmittedDataHandler))
	err := http.ListenAndServe("127.0.0.1:8886", nil)
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
		MemoryLimit := submittedData.MemoryLimit
		Memory := ""
		TimeLimit := submittedData.TimeLimit
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
				judgeResult := ""
				judgeResult, Time, Memory = JudgCmd(ProblemID, Username, Submitted, Lang, TimeLimit, MemoryLimit)

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
func JudgCmd(ProblemID string, Username string, Submitted string, Lang string, TimeLimit string, MemoryLimit string) (result string, costTime string, costMemory string) {
	sourceFilename, exeFilename, outFilename := GetFilename(ProblemID, Username, Submitted, Lang)
	stinputFilename, stoutputFilename := GetStioFilename(ProblemID)

	res := 0
	costMemory = ""
	start := time.Now()

	// CE RE TLE MLE
	// res = JudgeWithDocker(sourceFilename, exeFilename, outFilename, stinputFilename, Lang)

	res = JudgeCE(sourceFilename, exeFilename, Lang)
	if res != 0 {
		if res == 1 {
			return "Compile Error", "", ""
		} else { // System Error: [language error] or [Compile CMD Error]
			return "System Error", "", ""
		}
	}

	res = JudgeRE(exeFilename, outFilename, stinputFilename)
	if res != 0 {
		if res == 1 {
			return "Runtime Error", "", ""
		} else { // System Error: [Runtime CMD Error]
			return "System Error", "", ""
		}
	}

	duration := time.Since(start)
	durationMilliSeconds := duration.Nanoseconds()/1e6
	costTimeString := strconv.FormatInt(durationMilliSeconds, 10)
	costTimeInt, _ := strconv.Atoi(costTimeString)
	costTime = costTimeString+"MS"

	limitTimeArray := strings.Split(TimeLimit, "MS")
	limitTimeString := limitTimeArray[0]
	limitTimeInt, _ := strconv.Atoi(limitTimeString)
	// limitTimeInt -= 950
	
	if costTimeInt > limitTimeInt { // Time Limit Error
		// fmt.Println("!!!! costTimeInt:", costTimeInt, "...limitTimeInt:", limitTimeInt)
		return "Time Limit Error", "", ""
	}
	// fmt.Println("...costTimeString:", costTimeString, "...limitTimeString:", limitTimeString)
	

	
	res = JudgeWAandOLE(outFilename, stoutputFilename)
	if res != 0 {
		if res == 1 {
			return "Wrong Answer", "", ""
		} else if res == 2 {
			return "Output Limit Exceeded", "", ""
		} else { // System Error: [can't get file size] or [Wrong Answer CMD Error]
			return "System Error", "", ""
		}
	}

	res = JudgePE(outFilename, stoutputFilename)
	if res != 0 {
		if res == 1 {
			return "Presentation Error", "", ""
		} else { // System Error: [Presentation CMD Error]
			return "System Error", "", ""
		}
	}

	return "Accepted", costTime, costMemory
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
	// fmt.Println("#### Output Size: ", stoutputFileSize, " ... ", outFileSize)
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


/*
func JudgeWithDocker(sourceFilename string, exeFilename string, outFilename string, stinputFilename string, Lang string) (res int){

    ctx := context.Background()
    cli, err := client.NewEnvClient()
    defer cli.Close()
    if err != nil {
        panic(err)
    }
    checkAndStartContainer(ctx, cli)
	return 1;
}


//创建容器
func createContainer(ctx context.Context, cli *client.Client) {
    //创建容器
    cont, err := cli.ContainerCreate(ctx, &container.Config{
        Image:      imageName,     //镜像名称
        Tty:        true,          //docker run命令中的-t选项
        OpenStdin:  true,          //docker run命令中的-i选项
        Cmd:        []string{cmd}, //docker 容器中执行的命令
        WorkingDir: workDir,       //docker容器中的工作目录
        ExposedPorts: nat.PortSet{
            openPort: struct{}{}, //docker容器对外开放的端口
        },
    }, &container.HostConfig{
        PortBindings: nat.PortMap{
            openPort: []nat.PortBinding{nat.PortBinding{
                HostIP:   "0.0.0.0", //docker容器映射的宿主机的ip
                HostPort: hostPort,  //docker 容器映射到宿主机的端口
            }},
        },
        Mounts: []mount.Mount{ //docker 容器目录挂在到宿主机目录
            mount.Mount{
                Type:   mount.TypeBind,
                Source: hostDir,
                Target: containerDir,
            },
        },
    }, nil, containerName)
    if err == nil {
        log.Printf("success create container:%s\n", cont.ID)
    } else {
        log.Println("failed to create container!!!!!!!!!!!!!")
    }
}


//启动容器
func startContainer(ctx context.Context, containerID string, cli *client.Client) error {
    err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
    if err == nil {
        log.Printf("success start container:%s\n", containerID)
    } else {
        log.Printf("failed to start container:%s!!!!!!!!!!!!!\n", containerID)
    }
    return err
}
 
//将容器的标准输出输出到控制台中
func printConsole(ctx context.Context, cli *client.Client, id string) {
    //将容器的标准输出显示出来
    out, err := cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true})
    if err != nil {
        panic(err)
    }
    io.Copy(os.Stdout, out)
 
    //容器内部的运行状态
    status, err := cli.ContainerStats(ctx, id, true)
    if err != nil {
        panic(err)
    }
    io.Copy(os.Stdout, status.Body)
}
 
//检查容器是否存在并启动容器
func checkAndStartContainer(ctx context.Context, cli *client.Client) {
    for {
        select {
        case <-isRuning(ctx, cli):
            //该container没有在运行
            //获取所有的container查看该container是否存在
            contTemp := getContainer(ctx, cli, true)
            if contTemp.ID == "" {
                //该容器不存在，创建该容器
                log.Printf("the container name[%s] is not exists!!!!!!!!!!!!!\n", containerName)
                createContainer(ctx, cli)
            } else {
                //该容器存在，启动该容器
                log.Printf("the container name[%s] is exists\n", containerName)
                startContainer(ctx, contTemp.ID, cli)
            }
 
        }
    }
}
 
//获取container
func getContainer(ctx context.Context, cli *client.Client, all bool) types.Container {
    containerList, err := cli.ContainerList(ctx, types.ContainerListOptions{All: all})
    if err != nil {
        panic(err)
    }
    var contTemp types.Container
    //找出名为“mygin-latest”的container并将其存入contTemp中
    for _, v1 := range containerList {
        for _, v2 := range v1.Names {
            if v2 == indexName {
                contTemp = v1
                break
            }
        }
    }
    return contTemp
}
 
//容器是否正在运行
func isRuning(ctx context.Context, cli *client.Client) <-chan bool {
    isRun := make(chan bool)
    var timer *time.Ticker
    go func(ctx context.Context, cli *client.Client) {
        for {
            //每n s检查一次容器是否运行
 
            timer = time.NewTicker(time.Duration(n) * time.Second)
            select {
            case <-timer.C:
                //获取正在运行的container list
                log.Printf("%s is checking the container[%s]is Runing??", os.Args[0], containerName)
                contTemp := getContainer(ctx, cli, false)
                if contTemp.ID == "" {
                    log.Print(":NO")
                    //说明container没有运行
                    isRun <- true
                } else {
                    log.Print(":YES")
                    //说明该container正在运行
                    go printConsole(ctx, cli, contTemp.ID)
                }
            }
 
        }
    }(ctx, cli)
    return isRun
}
*/