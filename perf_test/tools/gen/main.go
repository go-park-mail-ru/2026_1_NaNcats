// Генератор целей (targets) для нагрузочного тестирования утилитой vegeta.
//
// Формирует JSONL-файл целей в формате vegeta (-format=json), по одной цели
// на строку. Поддерживает три режима:
//
//	create — POST /api/owner/restaurants (multipart/form-data, требует сессию owner
//	         + CSRF). Каждая цель уникальна: уникальное имя бренда и уникальный
//	         Idempotency-Key, поэтому N целей создают ровно N строк в БД.
//	search — GET /api/restaurants/search?q=... (публичный, бьёт по ILIKE-поиску).
//	list   — GET /api/restaurants/brands?limit&offset (публичный, deep offset).
//
// Запуск:
//
//	go run ./perf_test/tools/gen -mode=create -n=100000 -out=targets/create.json \
//	    -base=http://localhost:8080 -session=<id> -csrf=<token>
//	go run ./perf_test/tools/gen -mode=search -n=2000 -out=targets/search.json -base=...
//	go run ./perf_test/tools/gen -mode=list   -n=2000 -out=targets/list.json   -base=...
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"mime/multipart"
	"os"
)

// target — структура цели в JSON-формате vegeta. Поле Body сериализуется в
// base64 (стандартное поведение encoding/json для []byte), как и ожидает vegeta.
type target struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Body   []byte              `json:"body,omitempty"`
	Header map[string][]string `json:"header,omitempty"`
}

const boundary = "PERFBOUNDARY7a9c1e" // фиксированная граница multipart -> постоянный Content-Type

func main() {
	mode := flag.String("mode", "create", "create|search|list")
	n := flag.Int("n", 100000, "количество целей")
	out := flag.String("out", "targets.json", "путь к выходному файлу")
	base := flag.String("base", "http://localhost:8080", "базовый URL gateway")
	session := flag.String("session", "", "session_id для режима create")
	csrf := flag.String("csrf", "", "CSRF-токен для режима create")
	seed := flag.Int64("seed", 42, "seed ГПСЧ (для воспроизводимости)")
	maxOffset := flag.Int("max-offset", 99980, "верхняя граница offset для режима list")
	start := flag.Int("start", 0, "начальный индекс для режима create (для непересекающихся партий)")
	flag.Parse()

	rng := rand.New(rand.NewSource(*seed))

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot create out:", err)
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 1<<20)
	defer w.Flush()
	enc := json.NewEncoder(w)

	switch *mode {
	case "create":
		if *session == "" || *csrf == "" {
			fmt.Fprintln(os.Stderr, "create требует -session и -csrf")
			os.Exit(1)
		}
		for k := 0; k < *n; k++ {
			i := *start + k
			body := buildMultipart(i, rng)
			t := target{
				Method: "POST",
				URL:    *base + "/api/owner/restaurants",
				Body:   body,
				Header: map[string][]string{
					"Content-Type":    {"multipart/form-data; boundary=" + boundary},
					"Cookie":          {"session_id=" + *session},
					"X-CSRF-Token":    {*csrf},
					"Idempotency-Key": {fmt.Sprintf("perf-create-%08d", i)},
				},
			}
			if err := enc.Encode(&t); err != nil {
				fmt.Fprintln(os.Stderr, "encode:", err)
				os.Exit(1)
			}
		}
	case "search":
		// Набор поисковых токенов: и совпадающие с генерируемыми именами (perf, zeta),
		// и «промахи». Любой ILIKE '%q%' заставляет планировщик сканировать всю таблицу.
		// Токены длиной >= 3 символов — чтобы триграммный (pg_trgm) индекс мог
		// быть задействован планировщиком после оптимизации.
		tokens := []string{"perf", "rest", "zeta", "alpha", "burger", "pizza", "sushi",
			"xyz", "qwer", "0007", "brand", "777", "abc", "kzx", "rop"}
		for i := 0; i < *n; i++ {
			q := tokens[rng.Intn(len(tokens))]
			t := target{
				Method: "GET",
				URL:    fmt.Sprintf("%s/api/restaurants/search?q=%s&limit=20", *base, q),
			}
			if err := enc.Encode(&t); err != nil {
				fmt.Fprintln(os.Stderr, "encode:", err)
				os.Exit(1)
			}
		}
	case "list":
		for i := 0; i < *n; i++ {
			off := rng.Intn(*maxOffset + 1)
			t := target{
				Method: "GET",
				URL:    fmt.Sprintf("%s/api/restaurants/brands?limit=20&offset=%d", *base, off),
			}
			if err := enc.Encode(&t); err != nil {
				fmt.Fprintln(os.Stderr, "encode:", err)
				os.Exit(1)
			}
		}
	default:
		fmt.Fprintln(os.Stderr, "неизвестный режим:", *mode)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "сгенерировано %d целей в %s\n", *n, *out)
}

// buildMultipart собирает тело multipart/form-data для одной заявки на создание
// бренда. Имя уникально по индексу i (<= 60 символов), описание — случайное.
func buildMultipart(i int, rng *rand.Rand) []byte {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.SetBoundary(boundary)
	name := fmt.Sprintf("perf_%07d_%s", i, randStr(rng, 6))
	_ = writeField(mw, "name", name)
	_ = writeField(mw, "description", "perf load test brand "+randStr(rng, 40))
	_ = mw.Close()
	return buf.Bytes()
}

func writeField(mw *multipart.Writer, k, v string) error {
	fw, err := mw.CreateFormField(k)
	if err != nil {
		return err
	}
	_, err = fw.Write([]byte(v))
	return err
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randStr(rng *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return string(b)
}
