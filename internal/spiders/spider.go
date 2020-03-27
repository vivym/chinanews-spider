package spiders

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vivym/chinanews-spider/internal/wordle"

	"github.com/Kamva/mgm/v2"
	"github.com/vivym/chinanews-spider/internal/model"
	"go.mongodb.org/mongo-driver/mongo"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/PuerkitoBio/goquery"
	"github.com/panjf2000/ants/v2"

	"github.com/go-resty/resty/v2"
	"github.com/vivym/chinanews-spider/internal/nlp"
	"logur.dev/logur"
)

type Spider struct {
	config *Config
	logger logur.LoggerFacade
	nlp    *nlp.NLPToolkit
	wordle *wordle.WordleToolkit
}

type SpiderTask struct {
	task   Task
	config *Config
	logger logur.LoggerFacade
	nlp    *nlp.NLPToolkit
	wordle *wordle.WordleToolkit
	http   *resty.Client
}

func New(config Config, logger logur.LoggerFacade, nlpToolkit *nlp.NLPToolkit, wordleToolkit *wordle.WordleToolkit) *Spider {
	return &Spider{
		config: &config,
		logger: logger,
		nlp:    nlpToolkit,
		wordle: wordleToolkit,
	}
}

func (s *Spider) Run() int {
	var mu sync.Mutex
	sentences := make(map[string][]string)
	tags := make(map[string][]model.Tag)
	var wg sync.WaitGroup
	pool, _ := ants.NewPool(s.config.Concurrency)
	for key, tasks := range s.config.Tasks {
		for _, task := range tasks {
			if task.Skip {
				s.logger.Info("skip: " + task.Province)
				continue
			}

			s.logger.Info(task.Province)

			http := resty.New().
				SetRetryCount(3).
				SetRetryWaitTime(5*time.Second).
				SetRetryMaxWaitTime(20*time.Second).
				SetHostURL(task.BaseURL).
				SetHeader("User-Agent", userAgent).
				SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

			spiderTask := SpiderTask{
				config: s.config,
				task:   task,
				http:   http,
				nlp:    s.nlp,
				wordle: s.wordle,
				logger: s.logger,
			}

			region := key
			wg.Add(1)
			taskFunc := func() {
				sentence, tag := spiderTask.Run()
				mu.Lock()
				done := false
				sentences[region] = append(sentences[region], sentence)
				tags[region] = mergeTags(tags[region], tag)
				if len(sentences[region]) == len(tasks) {
					done = true
				}
				mu.Unlock()
				if done {
					sentence = strings.Join(sentences[region], "\n")
					if err := saveWords(region, sentence, tags[region], s.nlp, s.wordle); err != nil {
						s.logger.Error(err.Error())
					}
				}
				wg.Done()
			}
			_ = pool.Submit(taskFunc)
		}
	}

	wg.Wait()
	pool.Release()

	sentence := ""
	for _, s := range sentences {
		sentence += strings.Join(s, "\n")
	}
	var tag []model.Tag
	for _, t := range tags {
		tag = mergeTags(tag, t)
	}

	if err := saveWords("中国", sentence, tag, s.nlp, s.wordle); err != nil {
		s.logger.Error(err.Error())
	}

	return 0
}

func (s *SpiderTask) Run() (string, []model.Tag) {
	var sentences []string
	var tagsCounter = make(map[string]int)
	for _, subtask := range s.task.SubTasks {
		s.logger.Info(s.task.Province + " " + subtask.Tag)

		if subtask.Selector == nil {
			subtask.Selector = s.task.Selector
		}

		rsp, err := s.http.R().
			SetDoNotParseResponse(true).
			Get(subtask.URL)
		if err != nil {
			s.logger.Error("get news list http error: " + err.Error())
			return "", nil
		}

		body := rsp.RawBody()
		defer body.Close()
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			s.logger.Error("NewDocumentFromReader error: " + err.Error())
			return "", nil
		}

		isGBK := false
		doc.Find("meta").Each(func(_ int, el *goquery.Selection) {
			content, _ := el.Attr("content")
			if strings.Contains(content, "charset=gb") {
				isGBK = true
			}
			charset, _ := el.Attr("charset")
			if strings.Contains(charset, "gb") {
				isGBK = true
			}
		})

		repeatCnt := 0
		doc.Find(subtask.Selector.URL).Each(func(_ int, el *goquery.Selection) {
			if repeatCnt > 2 {
				return
			}
			url, _ := el.Find("a").Attr("href")
			if strings.HasPrefix(url, "/") || strings.HasPrefix(url, s.task.BaseURL) {
				delay := time.Duration(s.config.Delay + rand.Intn(200))
				time.Sleep(delay * time.Millisecond)

				sentence, ok := s.getNewsContent(url, s.task, subtask, isGBK)
				if !ok {
					repeatCnt++
				}
				if ok && len(sentence) > 0 {
					sentences = append(sentences, sentence)
					tagsCounter[subtask.Tag]++
				}
			}
		})
	}

	s.logger.Info(s.task.Province + " done. " + strconv.Itoa(len(sentences)))
	sentence := []rune(strings.Join(sentences, "\n"))
	limit := len(sentence)
	if limit > 2000000 {
		limit = 2000000
	}

	var tags []model.Tag
	for tag, count := range tagsCounter {
		tags = append(tags, model.Tag{
			Name:  tag,
			Count: count,
		})
	}

	if err := saveWords(s.task.Province, string(sentence[0:limit]), tags, s.nlp, s.wordle); err != nil {
		s.logger.Error(err.Error())
		return "", nil
	}

	limit = len(sentence)
	if limit > 200000 {
		limit = 200000
	}
	return string(sentence[0:limit]), tags
}

func saveWords(region, sentence string, tags []model.Tag, nlp *nlp.NLPToolkit, wordle *wordle.WordleToolkit) error {
	fmt.Println(region, len(sentence))
	keywords, err := nlp.ExtractKeywords(sentence, 300)
	if err != nil {
		return errors.New("ExtractKeywords error: " + err.Error())
	}

	if len(keywords) < 60 {
		return errors.New("not enough keywords: " + strconv.Itoa(len(keywords)))
	}

	news := model.News{
		Region: region,
		Date:   time.Now().Truncate(24 * time.Hour).Add(-8 * time.Hour),
	}

	news.Tags = tags

	news.Keywords, news.FillingWords, err = wordle.NCOVIS_ShapeWordle(keywords, region)
	if err != nil {
		return err
	}

	if err := mgm.Coll(&news).Create(&news); err != nil {
		return err
	}

	return nil
}

func (s *SpiderTask) getNewsContent(url string, task Task, subtask SubTask, isGBK bool) (string, bool) {
	rsp, err := s.http.R().
		SetDoNotParseResponse(true).
		Get(url)
	if err != nil {
		s.logger.Error("get news content http error: " + err.Error())
		return "", true
	}
	if rsp.StatusCode() != 200 {
		s.logger.Error("get news content http error: " + url + " " + rsp.Status())
		return "", true
	}

	body := rsp.RawBody()
	defer body.Close()

	var decodedBody io.Reader
	if isGBK {
		decodedBody = transform.NewReader(body, simplifiedchinese.GBK.NewDecoder())
	} else {
		decodedBody = body
	}
	doc, err := goquery.NewDocumentFromReader(decodedBody)
	if err != nil {
		s.logger.Error("NewDocumentFromReader error: " + err.Error())
		return "", true
	}

	if !strings.HasPrefix(url, "http") {
		url = task.BaseURL + url
	}

	title := strings.Trim(doc.Find(subtask.Selector.Title).Text(), " \n\r\t")
	src := strings.Trim(doc.Find(subtask.Selector.Src).Text(), " \n\r\t")
	content := strings.Trim(doc.Find(subtask.Selector.Content).Text(), " \n\r\t")

	if len(title) == 0 {
		s.logger.Error("can not find title " + url)
		return "", true
	}
	if len(src) == 0 {
		s.logger.Error("can not find src " + url)
		return "", true
	}

	var date time.Time
	date, err = parseDate(src)
	if err != nil {
		s.logger.Warn(err.Error() + " " + url + " isGBK: " + strconv.FormatBool(isGBK))
		return "", true
	}
	if date.Before(startDate) {
		return "", false
	}

	sentence := title + "\n" + content

	news := &model.NewsMetadata{
		Province: task.Province,
		URL:      url,
		Tag:      subtask.Tag,
		Title:    title,
		Date:     date,
	}
	if err := mgm.Coll(news).Create(news); err != nil {
		if merr, ok := err.(mongo.WriteException); ok {
			for _, werr := range merr.WriteErrors {
				if werr.Code != 11000 {
					s.logger.Error("insert news error: " + werr.Error())
				} else {
					return sentence, false
				}
			}
		} else {
			s.logger.Error("insert news unknown error: " + err.Error())
		}
	}

	return sentence, true
}
