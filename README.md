
## 🎯 目的

このチュートリアルでは、**Goテンプレート + htmx + Alpine.js** を使って以下を学びます。

- 初回レンダリング時にローディングを表示（`x-cloak` + `x-init`）
- 部分更新中にローディング表示（`.htmx-indicator`）
- Alpineの state 変更で特定アクションを実行（`x-effect` / `$watch`）
- Alpine の state 変更 → htmx の通信トリガ（`htmx.trigger`）
- htmx の DOM 差し替え後に Alpine を再初期化（`Alpine.initTree`）
- Next.js のようなページ遷移ローディング（`hx-boost` + グローバルバー）

---

## 📁 ディレクトリ構成

```text
htmx-alpine-tutorial/
├─ main.go
└─ templates/
   ├─ layout.html.tmpl
   ├─ index.html.tmpl
   └─ items.html.tmpl
```

---

## 前提

- Go がインストール済み（1.20 以降を想定）
- ターミナルで `go run` が使えること
- すべてローカル環境で完結する想定

ディレクトリは任意ですが、ここでは以下に作る想定です。

```bash
mkdir htmx-alpine-tutorial
cd htmx-alpine-tutorial
```

---

## STEP 0: 最小の HTTP サーバーを立ち上げる

まずは **Go の HTTP サーバーだけ** を動かします。

### 0-1. `main.go` を作成

`main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", helloHandler)

	log.Println("Listening on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, htmx + Alpine.js tutorial!")
}```

### 0-2. 実行して確認

```bash
go run main.go
```

ブラウザで <http://localhost:8080> を開き、<br/>
`Hello, htmx + Alpine.js tutorial!` と表示されれば OK です。

---

## STEP 1: Go テンプレートで HTML を返す

次に、**Go テンプレートを使って HTML を返す構成**に変更します。

### 1-1. テンプレート用ディレクトリを作成

```bash
mkdir templates
```

### 1-2. レイアウトテンプレートを作成（最初は素の HTML）

`templates/layout.html.tmpl`:

```html
<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <title>{{block "title" .}}{{.Title}}{{end}}</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
  {{block "body" .}}{{end}}
</body>
</html>
```

### 1-3. トップページテンプレートを作成

`templates/index.html.tmpl`:

```html
{{define "index.html.tmpl"}}
{{template "layout.html.tmpl" .}}

{{block "body" .}}
<main>
  <h1>HTMX + Alpine Tutorial</h1>
  <p>ここから少しずつ機能を足していきます。</p>
</main>
{{end}}

{{end}}
```

### 1-4. `main.go` をテンプレート対応に書き換え

`main.go` （上書き）:

```go
package main

import (
    "html/template"
    "log"
    "net/http"
)

func mustTemplates() *template.Template {
    // templates/ 以下の *.html.tmpl を全部読む
    t := template.Must(template.ParseGlob("templates/*.html.tmpl"))
    return t
}

func handleIndex(t *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        data := map[string]any{
            "Title": "HTMX + Alpine Tutorial",
        }
        if err := t.ExecuteTemplate(w, "index.html.tmpl", data); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}

func main() {
    t := mustTemplates()
    http.HandleFunc("/", handleIndex(t))

    log.Println("Listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 1-5. 動作確認

```bash
go run main.go
```

ブラウザで <http://localhost:8080> を開き、<br/>
`HTMX + Alpine Tutorial` の見出しが表示されれば OK です。

---

## STEP 2: Alpine.js で「初期ローディング画面」を追加する

ここでは **Alpine.js と `x-cloak` を使って、初期ローディング画面**を実装します。

### 2-1. レイアウトに Alpine.js と `x-cloak` のスタイルを追加

`templates/layout.html.tmpl` を次のように編集します（上書きでOK）:

```html
<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <title>{{block "title" .}}{{.Title}}{{end}}</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />

  <!-- Alpine.js 読み込み -->
  <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>

  <style>
    /* Alpine 起動まで x-cloak 要素は非表示 */
    [x-cloak]{ display:none !important; }
  </style>
</head>
<body>
  {{block "body" .}}{{end}}
</body>
</html>
```

### 2-2. `index.html.tmpl` にローディング表示を追加

`templates/index.html.tmpl` を以下のように編集します。

```html
{{define "index.html.tmpl"}}
{{template "layout.html.tmpl" .}}

{{block "body" .}}
<main x-data="{ ready:false }" x-init="ready = true">
  <!-- 初期ローディング（Alpine起動まで） -->
  <div x-show="!ready" x-cloak
       style="position:fixed;inset:0;display:grid;place-items:center;background:rgba(255,255,255,0.9);">
    <div style="width:32px;height:32px;border:4px solid #ddd;border-top-color:#555;border-radius:50%;animation:spin 1s linear infinite;"></div>
    <p style="margin-top:8px;color:#666;font-size:12px">Loading...</p>
  </div>

  <!-- 本文：Alpine起動後に表示 -->
  <section x-show="ready" x-cloak style="max-width:720px;margin:40px auto;padding:0 16px">
    <h1 style="font-size:24px;margin-bottom:12px">HTMX + Alpine Tutorial</h1>
    <p>ここから htmx と Alpine.js を組み合わせていきます。</p>
  </section>
</main>

<!-- 簡易スピンアニメーション -->
<style>
@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
{{end}}

{{end}}
```

### 2-3. 動作確認

```bash
go run main.go
```

リロードすると、一瞬「Loading...」が表示されてから本文が出れば成功です。

---

## STEP 3: htmx で「部分更新する /items リスト」を追加

ここから **htmx を導入して、ページ内の一部だけをサーバーから差し替える**処理を作ります。

### 3-1. レイアウトに htmx とインジケーター用スタイルを追加

`templates/layout.html.tmpl` を以下のように変更します（Alpine の行は残したまま追記）:

```html
<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <title>{{block "title" .}}{{.Title}}{{end}}</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />

  <!-- Alpine.js -->
  <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <!-- htmx -->
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>

  <style>
    [x-cloak]{ display:none !important; }

    /* htmx 通信中にだけ表示したいインジケーター用 */
    .htmx-indicator { display:none; }
    .htmx-request .htmx-indicator { display:inline-block; }
  </style>
</head>
<body>
  {{block "body" .}}{{end}}
</body>
</html>
```

### 3-2. `/items` 用テンプレートを作成

`templates/items.html.tmpl`:

```html
{{define "items.html.tmpl"}}
<ul style="margin-top:12px;display:grid;gap:6px">
  {{range .Items}}
    <li style="padding:8px;border:1px solid #eee;border-radius:6px">{{.}}</li>
  {{end}}
</ul>
{{end}}
```

### 3-3. `/items` エンドポイントを main.go に追加

`main.go` を **次のように差し替え**ます（既存関数に加筆・追加）。

```go
package main

import (
    "html/template"
    "log"
    "net/http"
    "strconv"
)

const pageSize = 5

func mustTemplates() *template.Template {
    t := template.Must(template.ParseGlob("templates/*.html.tmpl"))
    return t
}

func handleIndex(t *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        data := map[string]any{
            "Title": "HTMX + Alpine Tutorial",
        }
        if err := t.ExecuteTemplate(w, "index.html.tmpl", data); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}

// /items: リスト部分だけを返すハンドラ
func handleItems(t *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        page := 1
        if p := r.URL.Query().Get("page"); p != "" {
            if n, err := strconv.Atoi(p); err == nil && n > 0 {
                page = n
            }
        }

        start := (page - 1) * pageSize
        items := make([]string, 0, pageSize)
        for i := 0; i < pageSize; i++ {
            items = append(items, "Item #"+strconv.Itoa(start+i+1))
        }

        data := map[string]any{
            "Items": items,
            "Page":  page,
        }
        if err := t.ExecuteTemplate(w, "items.html.tmpl", data); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}

func main() {
    t := mustTemplates()
    http.HandleFunc("/", handleIndex(t))
    http.HandleFunc("/items", handleItems(t))

    log.Println("Listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 3-4. トップページから `/items` を htmx で読み込む

`templates/index.html.tmpl` の本文部分を、htmx 用の領域付きに変更します。

```html
{{define "index.html.tmpl"}}
{{template "layout.html.tmpl" .}}

{{block "body" .}}
<main x-data="{ ready:false }" x-init="ready = true">
  <!-- 初期ローディング（Alpine起動まで） -->
  <div x-show="!ready" x-cloak
       style="position:fixed;inset:0;display:grid;place-items:center;background:rgba(255,255,255,0.9);">
    <div style="width:32px;height:32px;border:4px solid #ddd;border-top-color:#555;border-radius:50%;animation:spin 1s linear infinite;"></div>
    <p style="margin-top:8px;color:#666;font-size:12px">Loading...</p>
  </div>

  <!-- 本文：Alpine起動後に表示 -->
  <section x-show="ready" x-cloak style="max-width:720px;margin:40px auto;padding:0 16px">
    <h1 style="font-size:24px;margin-bottom:12px">HTMX + Alpine Tutorial</h1>

    <p>下のリストは、htmx で <code>/items</code> から読み込みます。</p>

    <!-- htmx のロードインジケーター -->
    <span class="htmx-indicator" style="font-size:12px;color:#666;">読み込み中...</span>

    <!-- /items を読み込んで結果を差し替える領域 -->
    <div id="items"
         hx-get="/items?page=1"
         hx-trigger="load"
         hx-target="#items"
         hx-swap="innerHTML">
    </div>
  </section>
</main>

<style>
@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
{{end}}

{{end}}
```

### 3-5. 動作確認

```bash
go run main.go
```

- ページ表示後、**自動で `/items?page=1` にアクセス**してリストが描画される  
- 通信中は「読み込み中…」が表示される

ここまでで「**初期ローディング + htmx 部分更新**」まで完成です。

---

## STEP 4: Alpine の state 変更 → htmx 通信 という流れを作る

ここでは、**Alpine の `page` 状態が変わったら htmx で /items を再取得する**仕組みを追加します。

### 4-1. ページャー UI を追加（Alpine の state + `$watch`）

`templates/index.html.tmpl` の `<section>` 内を、次のように差し替えます。

```html
  <section x-show="ready" x-cloak style="max-width:720px;margin:40px auto;padding:0 16px">
    <h1 style="font-size:24px;margin-bottom:12px">HTMX + Alpine Tutorial</h1>

    <!-- 例: Alpine の state 変化 -->
    <div x-data="{ q:'', debounced:'' }"
         x-effect="
           clearTimeout($el._t);
           $el._t = setTimeout(() => debounced = q, 300)
         "
         style="padding:12px;border:1px solid #eee;border-radius:8px;margin-bottom:16px">
      <label style="display:block;font-size:12px;color:#666;margin-bottom:4px">クエリ（デバウンス例）</label>
      <input x-model="q" placeholder="type to debounce..."
             style="width:100%;padding:8px;border:1px solid #ddd;border-radius:6px">
      <div style="margin-top:6px;font-size:12px;color:#555">
        debounced: <span x-text="debounced"></span>
      </div>
    </div>

    <!-- ページャー: Alpine の page 変更 → htmx イベントで /items を再読込 -->
    <div x-data="{ page: 1 }"
         x-init="
           $watch('page', v => {
             // page が変わったらカスタムイベントを飛ばす
             htmx.trigger(document.body, 'items:reload', { detail: { page: v } })
           })
         "
         style="padding:12px;border:1px solid #eee;border-radius:8px">
      <div style="display:flex;align-items:center;gap:8px">
        <button type="button" @click="page = Math.max(1, page-1)">Prev</button>
        <span>Page: <strong x-text="page"></strong></span>
        <button type="button" @click="page++">Next</button>

        <!-- htmx 通信中インジケーター -->
        <span class="htmx-indicator" style="margin-left:auto;color:#666;font-size:12px">読み込み中…</span>
      </div>

      <!-- /items を読み込んで結果を差し替える領域 -->
      <div id="items"
           hx-get="/items?page=1"
           hx-trigger="load, items:reload from:body"
           hx-target="#items"
           hx-swap="innerHTML">
      </div>
    </div>
  </section>
```

### 4-2. カスタムイベントから page を取り出す（オプション）

今のままでも `/items?page=1` のままですが、<br/>
**イベントの detail に入っている page をクエリに反映したい場合**は、少し JavaScript を足します。

`templates/index.html.tmpl` の最後（`</main>` の直後あたり）に以下を追記します：

```html
<script>
  // items:reload で渡された page を、次回リクエスト時のクエリに反映する例
  document.body.addEventListener('items:reload', function (e) {
    const page = e.detail?.page || 1;
    const itemsDiv = document.getElementById('items');
    if (!itemsDiv) return;
    // hx-get の URL を動的に書き換える
    itemsDiv.setAttribute('hx-get', '/items?page=' + page);
  });
</script>
```

### 4-3. 動作確認

- ページロード時に 1 ページ目が表示される  
- Prev / Next を押すと `page` が変わり、  
  - `items:reload` イベントが飛ぶ  
  - `/items?page=N` にアクセスしてリストが更新される

---

## STEP 5: ページ遷移風のローディングバー（hx-boost）を追加

最後に、**Next.js のルートローディング風のバー**を追加します。<br/>
これは `hx-boost="true"` と htmx のグローバルイベントを使います。

### 5-1. レイアウトにローディングバーと hx-boost を追加

`templates/layout.html.tmpl` を次のように変更します。

```html
<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <title>{{block "title" .}}{{.Title}}{{end}}</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />

  <!-- Alpine.js -->
  <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <!-- htmx -->
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>

  <style>
    [x-cloak]{ display:none !important; }

    .htmx-indicator { display:none; }
    .htmx-request .htmx-indicator { display:inline-block; }

    /* グローバル進捗バー */
    #global-indicator{
      position:fixed;
      top:0;left:0;right:0;
      height:3px;
      background:#2563eb; /* 青 */
      transform:scaleX(0);
      transform-origin:left;
      transition:transform .2s ease;
      z-index:9999;
    }
  </style>
</head>
<body hx-boost="true">
  <!-- ページ遷移風プログレスバー -->
  <div id="global-indicator"></div>

  {{block "body" .}}{{end}}

  <script>
    const bar = document.getElementById('global-indicator');
    document.addEventListener('htmx:request', () => {
      bar.style.transform = 'scaleX(1)';
    });
    document.addEventListener('htmx:afterOnLoad', () => {
      setTimeout(() => {
        bar.style.transform = 'scaleX(0)';
      }, 150);
    });

    // 差し替えられたDOMにAlpineを再適用（htmx で入ってきた部分にも Alpine を効かせたい場合）
    document.addEventListener('htmx:afterSwap', (e) => {
      if (window.Alpine?.initTree) Alpine.initTree(e.target);
    });
  </script>
</body>
</html>
```

### 5-2. （任意）ページ内リンクで hx-boost を体験

今のチュートリアル構成だと `/` しかありませんが、<br/>
例えば `/about` などのページを生やして、`<a href="/about">` をクリックすると<br/>
**ページ全体を再描画する代わりに、htmx が中身だけ差し替え、その間上部のバーが伸びる**ようにできます。

（このチュートリアルでは割愛しますが、`hx-boost="true"` を `<body>` につけることで、<br/>
同一オリジンのリンク・フォーム送信が自動で AJAX 化されます）

---

## 最終的なファイル一覧

このチュートリアルを最後まで進めたとき、<br/>
最低限必要なファイルは次の 4 つです：

- `main.go`
- `templates/layout.html.tmpl`
- `templates/index.html.tmpl`
- `templates/items.html.tmpl`

それぞれの中身は、**STEP 3〜5 で示した最新のもの**になっていれば OK です。

---

## まとめ

- **STEP 0〜1**: Go の HTTP サーバー + テンプレートの骨組み
- **STEP 2**: Alpine.js の `x-data` / `x-init` / `x-cloak` で初期ローディング
- **STEP 3**: htmx の `hx-get` / `hx-trigger` / `hx-target` / `hx-swap` で部分更新
- **STEP 4**: Alpine の state (`page`) 変更を `$watch` で監視 → `htmx.trigger` で通信
- **STEP 5**: `hx-boost` + htmx グローバルイベントでルートローディング風バー

これで、**Next.js 的な「ロード感」を持った Go テンプレートベースの Web UI** を、<br/>
少しずつ段階を踏んで構築できるようになりました。
