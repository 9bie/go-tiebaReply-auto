package main
//贴吧自动抢楼.Golang
//既然你们说会被封号那么久叫封号器得了：：doge::
//http://www.9bie.org 本人小站 求围观qwq
//程序:贴吧自动回帖  语言:Golang
import "fmt"
import "net/http"
import "io/ioutil"
import "net/http/cookiejar"
import "regexp"
import "net/url"
import "bytes"
import "time"
import "os"
import "io"
import "strings"

//import "bufio"

var gCurCookies []*http.Cookie = nil
var gCurCookieJar *cookiejar.Jar
var tbs string           //贴吧发帖的一种验证
var Logined bool = false //作为判断可否发帖，因为要设置发帖时间
var dbPath string        //作为储存已经发送数据库的路径

//程序入口
func main() {
  gCurCookieJar, _ = cookiejar.New(nil) //Program Init
    dbPath = "text.txt"                   //DateBase Init
    if checkFileIsExist(dbPath) == false {
        os.Create(dbPath)
    }
    username, password, tieba := "", "", ""
    fmt.Print("请输入用户名:")
    fmt.Scanln(&username)
    fmt.Print("\n请输入密码:")
    fmt.Scanln(&password)
    loginTieba(username, password)
    fmt.Print("请输入要监控的贴吧:")
    fmt.Scanln(&tieba)
    controlTieba(tieba)

}

//获取贴吧的tbs和id
func tiebaInfo(tiebaname string) (string, string) {
    web := getUrlRespHtml("http://tieba.baidu.com/"+tiebaname, nil)
    tbs_temp, _ := regexp.Compile(`'tbs': "(?P<tbs_temp>\w+)"`)
    tbs := tbs_temp.FindStringSubmatch(web)
    gid, _ := regexp.Compile(`"forum_id":(?P<gid>\w+),"`)
    g_id := gid.FindStringSubmatch(web)
    return tbs[1], g_id[1]
}

/*
New List  <===>   Old list    ===>  delete
    ---go--->    (New List)
    new P_id ===> open file
                  Read Line    ====> check true/fasle
                                       true:Send Post{writed!}
                                            Write to File    => open timer(10min){writed?}
                                       }
*/
//获取贴吧帖子列表
func getTiebaPostList(tiebaname string) [50]string { //New List
    //tbs, gid := tiebaInfo(tiebaname)
    tb_List, _ := regexp.Compile(`<a href="(?P<tb_List>.+?)" title="(?:.+?)" target="_blank" class="j_th_tit ">`)
    list := tb_List.FindAllString(getUrlRespHtml("http://tieba.baidu.com/"+tiebaname, nil), -1)
    var postID [50]string
    for key, value := range list {
        //fmt.Print(value + "\n")
        postID[key] = tb_List.FindStringSubmatch(value)[1]
        //fmt.Print(postID[key])
    }
    return postID
}

//判断文件是否存在
func checkFileIsExist(filename string) bool {
    var exist = true
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        exist = false
    }
    return exist
}

//把新帖PID写入数据库
func writeNewPidToData(new_Pid string) {
    file, _ := os.OpenFile(dbPath, os.O_APPEND, 0666)
    defer file.Close()
    io.WriteString(file, new_Pid+"|")
}

//检查新帖是否在文本文件的数据库中
func checkFilePostIdData(new_pid string) bool {
    b, _ := ioutil.ReadFile(dbPath)
    s := string(b)
    for _, lineStr := range strings.Split(s, "|") {
        if lineStr == new_pid {
            return true
        } else {
            continue
        }
    }
    return false
}

//检测
func checkList(old_list [50]string, tiebaname string) (new_list [50]string, new_Pid []string) {
    //下次传进来old_list的就是上一次本函数输出的new_list,new_Pid是新贴列表，待回复的
    length_old_list := len(old_list)
    new_list = getTiebaPostList(tiebaname)
    for _, value_new := range new_list {

        temp := value_new //遍历取出B中的元素

        for j := 0; j < length_old_list; j++ {
            if temp == old_list[j] { //如果相同 比较下一个
                break
            } else {
                if length_old_list == (j + 1) { //如果不同 查看a的元素个数及当前比较元素的位置 将不同的元素添加到返回slice中
                    new_Pid = append(new_Pid, temp)
                    //fmt.Println("---->", new_Pid)
                }
            }
        }
    }
    return new_list, new_Pid
}

//监控贴吧
func controlTieba(tiebaName string) {
    New_list := getTiebaPostList(tiebaName)
    tbs, gid := tiebaInfo(tiebaName)
    fmt.Print("获取到的tbs为:", tbs, "贴吧id为:", gid)
    for {
        var new_Pid []string
        New_list, new_Pid = checkList(New_list, tiebaName)
        for _, value := range new_Pid {
            if checkFilePostIdData(value) == false {
                fmt.Print("新帖子:", value, "\n")
                if Logined == true {
                    postid := strings.Split(value, "/")
                    fmt.Print("正在发送:", postid[2], "\n")
                    tiebaReply(tiebaName, gid, postid[2], "前排围观\r\n本条信息由程序发送，本程序由GOLANG语言编写qwq，如果感到厌烦请回复，我将关闭此程序", tbs)
                    writeNewPidToData(postid[2])
                } else {
                    fmt.Print("发现有新帖子，请先登录\n")
                }

            }
        }
        time.Sleep(30 * time.Second)
        fmt.Print("Looping.....\n")
    }
}

//贴吧回帖
func tiebaReply(tiebaname string, tiebaid string, postid string, text string, tbs string) {
    getUrl := "http://tieba.baidu.com/f/commit/post/add"
    //name, _ := url.Parse(tiebaname)
    //txt, _ := url.Parse(text)
    var getData map[string]string
    getData = make(map[string]string)
    getData["kw"] = tiebaname
    getData["rich_text"] = "1"
    getData["floor_num"] = "1"
    getData["ie"] = "utf-8"
    getData["fid"] = tiebaid
    getData["tid"] = postid
    getData["content"] = text
    getData["tbs"] = tbs
    getData["files"] = "[]"
    getData["mouse_pwd_t"] = string(time.Now().Unix())
    getData["mouse_pwd_isclick"] = "0"
    getData["__type__"] = "reply"
    fmt.Print(getUrlRespHtml(getUrl, getData), "\n")
}

//Header Data
func newPostData() map[string]string {
    var postDict map[string]string
    postDict = make(map[string]string)
    postDict["Host"] = "passport.baidu.com"
    postDict["User-Agent"] = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:27.0) Gecko/20100101 Firefox/27.0"
    postDict["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
    postDict["Accept-Language"] = "en-US,en;q=0.5'"
    postDict["Host"] = "passport.baidu.com"
    postDict["Accept-Encoding"] = "gzip, deflate"
    postDict["Refer"] = "http://pan.baidu.com/"
    postDict["Content-Type"] = "application/x-www-form-urlencoded"
    return postDict
}

//get url response html
func getUrlRespHtml(strUrl string, postDict map[string]string) string {
    var respHtml string = ""
    httpClient := &http.Client{
        Jar: gCurCookieJar,
    }

    var httpReq *http.Request
    //var newReqErr error
    if nil == postDict {
        httpReq, _ = http.NewRequest("GET", strUrl, nil)
    } else {
        postValues := url.Values{}
        for postKey, PostValue := range postDict {
            postValues.Set(postKey, PostValue)
        }
        postDataStr := postValues.Encode()
        postDataBytes := []byte(postDataStr)
        postBytesReader := bytes.NewReader(postDataBytes)
        httpReq, _ = http.NewRequest("POST", strUrl, postBytesReader)
        httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    }
    httpResp, err := httpClient.Do(httpReq)
    if err != nil {
    }
    defer httpResp.Body.Close()
    body, errReadAll := ioutil.ReadAll(httpResp.Body)
    if errReadAll != nil {
    }
    gCurCookies = gCurCookieJar.Cookies(httpReq.URL)
    respHtml = string(body)
    return respHtml
}

//debug
func dbgPrintCurCookies() {
    var cookieNum int = len(gCurCookies)
    for i := 0; i < cookieNum; i++ {
        var curCk *http.Cookie = gCurCookies[i]
        //{Name,Value,Path,Domain,Expires,RawExpires,MaxAge,Secure,Httponly,Raw,Unparsed}
        fmt.Print("Name=%s\n", curCk.Name)
        fmt.Print("Value=%s\n", curCk.Value)
        fmt.Print("Path=%s\n", curCk.Path)
    }
}

//login baidu
func loginTieba(username string, password string) {
    getUrlRespHtml("http://pan.baidu.com", nil)
    token := getUrlRespHtml("https://passport.baidu.com/v2/api/?getapi&class=login&tpl=mn&tangram=true", nil)
    loginTokenP, _ := regexp.Compile(`bdPass\.api\.params\.login_token='(?P<loginToken>\w+)';`)
    loginToken := loginTokenP.FindStringSubmatch(token)
    var postData map[string]string
    postData = newPostData()
    const loginUrl string = "https://passport.baidu.com/v2/api/?login"
    //构造第一次登陆包
    postData["staticpage"] = "http://pan.baidu.com/res/static/thirdparty/pass_v3_jump.html"
    postData["charset"] = "utf-8"
    postData["token"] = loginToken[1]
    postData["tpl"] = "netdisk"
    postData["apiver"] = "v3"
    postData["tt"] = string(time.Now().Unix())
    postData["codestring"] = ""
    postData["safeflg"] = "0"
    postData["u"] = "http://pan.baidu.com"
    postData["isPhone"] = "false"
    postData["quick_user"] = "0"
    postData["loginmerge"] = "true"
    postData["logintype"] = "basicLogin"
    postData["username"] = username
    postData["password"] = password
    postData["verifycode"] = ""
    postData["mem_pass"] = "on"
    postData["ppui_logintime"] = "49586"
    postData["callback"] = "parent.bd__pcbs__hksq59"
    //第一次登陆包构建完成
    login := getUrlRespHtml(loginUrl, postData)
    verifcodeP, _ := regexp.Compile(`&codeString=(?P<verifcodeP>\w+)&`)
    verifcode_url := verifcodeP.FindStringSubmatch(login)
    fmt.Print("请打开链接读取图片并写下验证码:", "https://passport.baidu.com/cgi-bin/genimage?"+verifcode_url[1], "\n验证码：")
    writeNewPidToData("https://passport.baidu.com/cgi-bin/genimage?" + verifcode_url[1])
    verifycode := ""
    fmt.Scanln(&verifycode)
    //开始构建第二次登陆
    postData["token"] = loginToken[1]
    postData["tt"] = string(time.Now().Unix())
    postData["codestring"] = verifcode_url[1]
    postData["username"] = username
    postData["password"] = password
    postData["verifycode"] = verifycode
    delete(postData, "mem_pass")
    getUrlRespHtml(loginUrl, postData)
    //第二次构造登陆完成
    Logined = true
}