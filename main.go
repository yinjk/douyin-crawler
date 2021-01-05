/**
 *
 * @author yinjk
 * @create 2019-04-04 17:14
 */
package main

import (
	"douyin-crawler/httpclient"
	"flag"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type AwemeList struct {
	StatusCode int     `json:"status_code"`
	HasMore    int     `json:"has_more"`
	MinCursor  int     `json:"min_cursor"`
	MaxCursor  int     `json:"max_cursor"`
	AwemeList  []Aweme `json:"aweme_list"`
}

type Aweme struct {
	AwemeId   int64                    `json:"aweme_id"`
	Author    map[string]interface{}   `json:"author"`
	AwemeType int                      `json:"aweme_type"`
	ChaList   []map[string]interface{} `json:"cha_list"`
	Desc      string                   `json:"desc"`
	Music     map[string]interface{}   `json:"music"`
	Video     Video                    `json:"video"`
}
type Video struct {
	PlayAddr     UrlList `json:"play_addr"`
	Cover        UrlList `json:"cover"`
	DynamicCover UrlList `json:"dynamic_cover"`
}
type UrlList struct {
	UrlList []string `json:"url_list"`
}

type Data struct {
	Cursor     int     `json:"cursor"`
	StatusCode int     `json:"status_code"`
	HasMore    int     `json:"has_more"`
	AwemeList  []Aweme `json:"aweme_list"`
}

var userList = []string{"67356676104", "77716677707"}
var chMap = map[string]string{"jk": "1566807698522114", "小姐姐泳衣": "1608475782954061"}

const defaultRootPath = "F://download/"

var (
	h, v bool

	cursor, count string

	chId, rootPath string
)

func init() {
	flag.BoolVar(&h, "help", false, "this help")
	flag.BoolVar(&v, "version", false, "show version and exit")

	flag.StringVar(&cursor, "cursor", "0", "the offset of the video list")
	flag.StringVar(&count, "count", "20", "the total of the video will to search")

	flag.StringVar(&chId, "c", "", "the challenge id")
	flag.StringVar(&rootPath, "path", defaultRootPath, "the rootPath for video download, and the default is F://download/")

	// 改变默认的 Usage
	flag.Usage = usage
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}
	if v {
		version()
		return
	}

	fmt.Println("开始解析url...")
	urls := getChallengeDownloadUrls(cursor, count, 10) //重试10次
	if urls == nil || len(urls) == 0 {
		fmt.Println("下载失败，可能原因：1.抖音更改了签名算法。2.签名生成异常")
		return
	}
	fmt.Println("搜索到视屏数量: ", len(urls))
	directoryName := chId + "@" + cursor + "~" + count + "-" + getTime()
	downPath := path.Join(rootPath, directoryName)
	downloadTask(urls, downPath)

}

func downloadTask(urls []string, downPath string) {
	var (
		successCount = 0
		failCount    = 0
		total        = len(urls)
	)
	fmt.Println("下载任务开始，视屏将被下载到目录：" + downPath + " ,请耐心等待...")
	for _, v := range urls {
		if err := download(v, downPath); err != nil {
			failCount++
			fmt.Println("[下载失败]: ", v, " || 失败原因:", err.Error())
		} else {
			successCount++
			fmt.Println("[下载成功]: ", v, fmt.Sprintf(".....当前进度 [%d/%d] _ (%.2f%%)", successCount+failCount, total, float64(100*(successCount+failCount))/float64(total)))
		}
	}
	fmt.Println("=======================下载完成======================")
	fmt.Println(fmt.Sprintf("           [成功:%d] - [失败:%d] - [总数:%d]        ", successCount, failCount, total))
	fmt.Println("====================================================")
}

func getTime() string {
	now := time.Now()
	return now.Format("2006-01-02-150405")
}

func getChallengeDownloadUrls(cursor, count string, retry int) (urls []string) {
	url := `https://www.iesdouyin.com/aweme/v1/challenge/aweme/`
	if chId == "" {
		chId = chMap["jk"]
	}
	var data Data
	signature := GetSignature()
	value := httpclient.NewFormValue()
	value.Set("ch_id", chId)
	value.Set("cursor", cursor)
	value.Set("count", count)
	value.Set("aid", "1128")
	value.Set("screen_limit", "3")
	value.Set("download_click_limit", "0")
	value.Set("_signature", signature)

	_, _ = httpclient.GetParam(url, value, &data)
	for i := 0; (data.AwemeList == nil || len(data.AwemeList) == 0) && i < 100; i++ {
		_, _ = httpclient.GetParam(url, value, &data)
	}

	if data.AwemeList == nil || len(data.AwemeList) == 0 { //获取下载地址失败
		if retry > 0 { //失败重试
			retry--
			fmt.Println("解析url失败，正在重试，剩余重试次数：", retry)
			return getChallengeDownloadUrls(cursor, count, retry)
		}
		return
	}
	fmt.Println("url解析成功，正在获取视屏下载地址...")
	for _, v := range data.AwemeList {
		urlList := v.Video.PlayAddr.UrlList
		if urlList != nil && len(urlList) > 0 {
			urls = append(urls, urlList[0])
		}
	}
	return
}

func download(url string, filePath string) (err error) {
	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Host", "aweme.snssdk.com")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusFound {
		trueUrl := resp.Header["Location"][0]
		resp, err = http.Get(trueUrl)
		if err != nil {
			return err
		}
	}
	fileName := strings.Replace(url, "https://aweme.snssdk.com/aweme/v1/playwm/?video_id=", "", -1)
	fileName = strings.ReplaceAll(fileName, "&line=0", "")
	fileName = fileName + ".mp4"
	if pathExists, err := isPathExists(filePath); err != nil {
		return err
	} else if !pathExists {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
	}
	f, err := os.Create(path.Join(filePath, fileName))
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	return err
}

func isPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetSignature() (signature string) {
	userId := userList[0]
	vm := goja.New()
	vm.Set("userId", userId)
	_, e := vm.RunScript("", genSignature)
	if e != nil {
		panic(e)
	}
	return vm.Get("signature").String()
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `douyin-crawler version: 0.0.1
Usage: douyin-crawler [-v] [-b begin] [-t total] [-c challengeId] [-d rootPath]

Options:
`)
	flag.PrintDefaults()
}

func version() {
	_, _ = fmt.Fprintf(os.Stderr, `douyin-crawler version: 0.0.1`)
}

var genSignature = `
    this.navigator = {
        userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1"
    }
    var e = {}

    var r = (function () {
        function e(e, a, r) {
            return (b[e] || (b[e] = t("x,y", "return x " + e + " y")))(r, a)
        }

        function a(e, a, r) {
            return (k[r] || (k[r] = t("x,y", "return new x[y](" + Array(r + 1).join(",x[++y]").substr(1) + ")")))(e, a)
        }

        function r(e, a, r) {
            var n, t, s = {}, b = s.d = r ? r.d + 1 : 0;
            for (s["$" + b] = s, t = 0; t < b; t++) s[n = "$" + t] = r[n];
            for (t = 0, b = s.length = a.length; t < b; t++) s[t] = a[t];
            return c(e, 0, s)
        }

        function c(t, b, k) {
            function u(e) {
                v[x++] = e
            }

            function f() {
                return g = t.charCodeAt(b++) - 32, t.substring(b, b += g)
            }

            function l() {
                try {
                    y = c(t, b, k)
                } catch (e) {
                    h = e, y = l
                }
            }

            for (var h, y, d, g, v = [], x = 0; ;) switch (g = t.charCodeAt(b++) - 32) {
                case 1:
                    u(!v[--x]);
                    break;
                case 4:
                    v[x++] = f();
                    break;
                case 5:
                    u(function (e) {
                        var a = 0, r = e.length;
                        return function () {
                            var c = a < r;
                            return c && u(e[a++]), c
                        }
                    }(v[--x]));
                    break;
                case 6:
                    y = v[--x], u(v[--x](y));
                    break;
                case 8:
                    if (g = t.charCodeAt(b++) - 32, l(), b += g, g = t.charCodeAt(b++) - 32, y === c) b += g; else if (y !== l) return y;
                    break;
                case 9:
                    v[x++] = c;
                    break;
                case 10:
                    u(s(v[--x]));
                    break;
                case 11:
                    y = v[--x], u(v[--x] + y);
                    break;
                case 12:
                    for (y = f(), d = [], g = 0; g < y.length; g++) d[g] = y.charCodeAt(g) ^ g + y.length;
                    u(String.fromCharCode.apply(null, d));
                    break;
                case 13:
                    y = v[--x], h = delete v[--x][y];
                    break;
                case 14:
                    v[x++] = t.charCodeAt(b++) - 32;
                    break;
                case 59:
                    u((g = t.charCodeAt(b++) - 32) ? (y = x, v.slice(x -= g, y)) : []);
                    break;
                case 61:
                    u(v[--x][t.charCodeAt(b++) - 32]);
                    break;
                case 62:
                    g = v[--x], k[0] = 65599 * k[0] + k[1].charCodeAt(g) >>> 0;
                    break;
                case 65:
                    h = v[--x], y = v[--x], v[--x][y] = h;
                    break;
                case 66:
                    u(e(t[b++], v[--x], v[--x]));
                    break;
                case 67:
                    y = v[--x], d = v[--x], u((g = v[--x]).x === c ? r(g.y, y, k) : g.apply(d, y));
                    break;
                case 68:
                    u(e((g = t[b++]) < "<" ? (b--, f()) : g + g, v[--x], v[--x]));
                    break;
                case 70:
                    u(!1);
                    break;
                case 71:
                    v[x++] = n;
                    break;
                case 72:
                    v[x++] = +f();
                    break;
                case 73:
                    u(parseInt(f(), 36));
                    break;
                case 75:
                    if (v[--x]) {
                        b++;
                        break
                    }
                case 74:
                    g = t.charCodeAt(b++) - 32 << 16 >> 16, b += g;
                    break;
                case 76:
                    u(k[t.charCodeAt(b++) - 32]);
                    break;
                case 77:
                    y = v[--x], u(v[--x][y]);
                    break;
                case 78:
                    g = t.charCodeAt(b++) - 32, u(a(v, x -= g + 1, g));
                    break;
                case 79:
                    g = t.charCodeAt(b++) - 32, u(k["$" + g]);
                    break;
                case 81:
                    h = v[--x], v[--x][f()] = h;
                    break;
                case 82:
                    u(v[--x][f()]);
                    break;
                case 83:
                    h = v[--x], k[t.charCodeAt(b++) - 32] = h;
                    break;
                case 84:
                    v[x++] = !0;
                    break;
                case 85:
                    v[x++] = void 0;
                    break;
                case 86:
                    u(v[x - 1]);
                    break;
                case 88:
                    h = v[--x], y = v[--x], v[x++] = h, v[x++] = y;
                    break;
                case 89:
                    u(function () {
                        function e() {
                            return r(e.y, arguments, k)
                        }

                        return e.y = f(), e.x = c, e
                    }());
                    break;
                case 90:
                    v[x++] = null;
                    break;
                case 91:
                    v[x++] = h;
                    break;
                case 93:
                    h = v[--x];
                    break;
                case 0:
                    return v[--x];
                default:
                    u((g << 16 >> 16) - 16)
            }
        }

        var n = this, t = n.Function, s = Object.keys || function (e) {
            var a = {}, r = 0;
            for (var c in e) a[r++] = c;
            return a.length = r, a
        }, b = {}, k = {};
        return r
    })()('gr$Daten Иb/s!l y͒yĹg,(lfi~ah` + "`" + `{mv,-n|jqewVxp{rvmmx,&effkx[!cs"l".Pq%widthl"@q&heightl"vr*getContextx$"2d[!cs#l#,*;?|u.|uc{uq$fontl#vr(fillTextx$$龘ฑภ경2<[#c}l#2q*shadowBlurl#1q-shadowOffsetXl#$$limeq+shadowColorl#vr#arcx88802[%c}l#vr&strokex[ c}l"v,)}eOmyoZB]mx[ cs!0s$l$Pb<k7l l!r&lengthb%^l$1+s$jl  s#i$1ek1s$gr#tack4)zgr#tac$! +0o![#cj?o ]!l$b%s"o ]!l"l$b*b^0d#>>>s!0s%yA0s"l"l!r&lengthb<k+l"^l"1+s"jl  s&l&z0l!$ +["cs\'(0l#i\'1ps9wxb&s() &{s)/s(gr&Stringr,fromCharCodes)0s*yWl ._b&s o!])l l Jb<k$.aj;l .Tb<k$.gj/l .^b<k&i"-4j!+& s+yPo!]+s!l!l Hd>&l!l Bd>&+l!l <d>&+l!l 6d>&+l!l &+ s,y=o!o!]/q"13o!l q"10o!],l 2d>& s.{s-yMo!o!]0q"13o!]*Ld<l 4d#>>>b|s!o!l q"10o!],l!& s/yIo!o!].q"13o!],o!]*Jd<l 6d#>>>b|&o!]+l &+ s0l-l!&l-l!i\'1z141z4b/@d<l"b|&+l-l(l!b^&+l-l&zl\'g,)gk}ejo{cm,)|yn~Lij~em["cl$b%@d<l&zl\'l $ +["cl$b%b|&+l-l%8d<@b|l!b^&+ q$sign ', [e])
    var signature = e.sign(userId)
`
