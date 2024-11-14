
I dati per i comuni si scaricano come nelle snippets seguenti, ma è meglio inserire
i file come embed poiché sono file statici che non variano mai

```go
r, err := http.Get("https://github.com/opendatasicilia/comuni-italiani/raw/main/dati/main.csv")
if err != nil {
	t.Fatal(err)
}
defer r.Body.Close()
b, _ := io.ReadAll(r.Body)
err = os.WriteFile("comuni.csv", b, 0666)
```


```go
r, err := http.Get("https://github.com/opendatasicilia/comuni-italiani/raw/main/dati/popolazione_2021.csv")
if err != nil {
	t.Fatal(err)
}
defer r.Body.Close()
b, _ := io.ReadAll(r.Body)
err = os.WriteFile("popolazione_2021.csv", b, 0666)
```