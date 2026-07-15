# Pacote Pacific

O pacote `pacific` fornece um cliente HTTP para o sistema Pacific: construção
do payload de entrada (`PacificInput`), envio via `PacificHttpRepository`, e
detecção de respostas de erro (que o Pacific às vezes retorna com HTTP 200).

## Estruturas Principais

### `PacificInput`
```go
type PacificInput struct {
    UID    string  `json:"uid"`
    PWD    string  `json:"pwd"`
    PtoP   string  `json:"ptoP"`
    Params []Param `json:"params"`
}

type Param struct {
    Parametro string `json:"parametro"`
    DataType  string `json:"data_type"`
    ParamType string `json:"param_type"`
    Valor     string `json:"valor"`
}
```

`UID`/`PWD` são as credenciais do usuário Pacific, `PtoP` identifica o
programa/serviço a ser chamado. `Params` sempre inclui `pcMetodo` (o método a
executar), `pcParametros` (o payload, tipicamente um JSON serializado como
string) e `pcRetorno` (parâmetro de saída, preenchido pelo Pacific).

### `Dados`
```go
type Dados struct {
    Usuario  string
    Senha    string
    Programa string
    Metodo   string
    Valor    string
    IsGed    bool
}
```

Struct de conveniência para agrupar os dados de uma chamada antes de montar o
`PacificInput` — não é serializada diretamente, é só um agrupador de
parâmetros para os construtores abaixo.

## Construindo um `PacificInput`

```go
input := pacific.NewPacificInput(usuario, senha, programa, metodo, valorJSON)
```

Para operações que exigem também uma senha de colaborador (ex.: aprovações,
operações administrativas):
```go
input := pacific.NewPacificInputColab(usuario, senha, programa, metodo, valorJSON, senhaColaborador)
```

## Enviando a Requisição

```go
repo := pacific.NewPacificHttpRepository()

body, err := repo.Send(ctx, "https://pacific.example.com/api", input, 30*time.Second)
if err != nil {
    var pacificErr pacific.PacificError
    if errors.As(err, &pacificErr) {
        log.Printf("pacific falhou (status %d): %v", pacificErr.StatusCode, pacificErr)
    }
    return err
}
```

`Send` faz um `PUT` com o `PacificInput` serializado em JSON, aplica o
`timeout` recebido via `context.WithTimeout`, e retorna erro tanto para
falhas de transporte/HTTP quanto para respostas que o Pacific retorna com
status 200 mas que contêm um erro de aplicação no corpo (ver `IsResponseErr`
abaixo).

### Customizando o repository

`NewPacificHttpRepository` aceita opções funcionais:

```go
repo := pacific.NewPacificHttpRepository(
    pacific.WithLogger(myLogger),      // *slog.Logger — padrão: slog.Default()
    pacific.WithHTTPClient(myClient),  // *http.Client — padrão: cliente dedicado do pacote
)
```

- **`WithLogger`**: injeta um `*slog.Logger` próprio (por exemplo, com
  atributos fixos de serviço/ambiente). O repository nunca loga `UID`, `PWD`
  ou o conteúdo de `Params` — apenas `url`, `programa` (`PtoP`), status HTTP e
  duração da chamada, em nível `debug` (início da chamada), `info` (sucesso)
  ou `error` (falha).
- **`WithHTTPClient`**: substitui o `*http.Client` usado nas chamadas. Por
  padrão, o repository cria seu **próprio** `*http.Client`/`*http.Transport`
  com `InsecureSkipVerify: true` — isolado do `http.DefaultClient`/
  `http.DefaultTransport` do processo, para não afetar (nem sofrer
  interferência de) qualquer outro código HTTP rodando no mesmo binário.

## Tratamento de Erros

### `PacificError`
```go
type PacificError struct {
    StatusCode int
    Body       []byte
    Message    string
    Err        error // erro original de rede/parse, quando houver
}
```

Implementa `error` e `Unwrap() error`, então `errors.Is`/`errors.As` funcionam
normalmente sobre o erro original (timeout, falha de DNS, etc.) quando
presente.

### Detectando erro de aplicação em resposta HTTP 200

O Pacific às vezes retorna HTTP 200 com um corpo indicando erro. `IsResponseErr`
verifica os dois formatos conhecidos:

```go
type LogErroApp struct {
    LogErroApp []LogErroAppElement `json:"logErroApp"`
}

type LogErroAppElement struct {
    ID   int64  `json:"id"`
    Erro string `json:"erro"`
}

type LogErr001 struct {
    Status string `json:"status"`
    Msg    string `json:"msg"`
}

if pacific.IsResponseErr(body) {
    // corpo contém logErroApp[0].erro não vazio, ou status == "ERRO"
}
```

`Send` já chama `IsResponseErr` internamente e retorna `PacificError` quando
aplicável — normalmente não é necessário chamar diretamente, exceto ao
inspecionar uma resposta já obtida por outro caminho.

## Segurança

- Nunca logue `input.UID`/`input.PWD` nem o conteúdo de `input.Params` — o
  logging embutido no repository já evita isso por padrão; mantenha a mesma
  disciplina em qualquer log adicional que seu código adicionar em torno de
  `PacificInput`.
- O cliente HTTP padrão usa `InsecureSkipVerify: true` (compatibilidade com
  os endpoints Pacific atuais). Se o ambiente de destino tiver um certificado
  válido, prefira substituir por um cliente com verificação de TLS habilitada
  via `WithHTTPClient`.

## Dependências

- `encoding/json`, `net/http`, `crypto/tls`, `log/slog`, `context`, `time` — apenas biblioteca padrão do Go.

## Veja Também

- [Pacote Auth](../auth/README.md) - Para autenticação
- [Pacote Formatter](../formatter/README.md) - Para formatação de erros
- [Pacote Validator](../validator/README.md) - Para validação de dados

---

**Nota**: Este pacote é específico para integração com sistemas Pacific.
Certifique-se de ter as credenciais e permissões adequadas antes de usar em
produção.
