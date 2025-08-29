# Pacote GCS

O pacote `gcs` fornece uma função de inicialização simplificada para o cliente do Google Cloud Storage (GCS), facilitando a configuração e autenticação para interagir com buckets e objetos.

## Funcionalidades

### 🔧 Inicialização Simplificada

- Configuração automática do cliente GCS
- Autenticação via arquivo de credenciais
- Retorna cliente pronto para uso
- Gerenciamento de contexto

### 🔒 Autenticação

- Suporte a Service Account
- Carregamento de credenciais via arquivo JSON
- Configuração segura de autenticação

## Estruturas Principais

### `GCSConfig`

```go
type GCSConfig struct {
    *storage.Client
}
```

Wrapper que encapsula o cliente do Google Cloud Storage.

### `GCSClientConfig`

```go
type GCSClientConfig struct {
    context  context.Context
    filePath string
}
```

Configuração interna para inicialização do cliente.

### `GCSClientOption`

```go
type GCSClientOption func(d *GCSClientConfig)
```

Tipo de função para opções de configuração (extensibilidade futura).

## Configuração

### Inicialização Básica

```go
package main

import (
    "context"
    "log"
    "seu-projeto/initializers/gcs"
)

func main() {
    ctx := context.Background()
    credentialsPath := "/path/to/service-account.json"

    // Inicializar cliente GCS usando o pacote
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    log.Println("Cliente GCS inicializado com sucesso")
}
```

### Configuração via Variáveis de Ambiente

```go
package main

import (
    "context"
    "log"
    "os"
    "seu-projeto/initializers/gcs"
)

func setupGCSFromEnv() {
    ctx := context.Background()

    // Obter caminho das credenciais via variável de ambiente
    credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
    if credentialsPath == "" {
        log.Fatal("GOOGLE_APPLICATION_CREDENTIALS não definida")
    }

    // Inicializar cliente
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    log.Println("Cliente GCS configurado via variáveis de ambiente")
}
```

## Uso Básico

### Listagem de Buckets

```go
package main

import (
    "context"
    "fmt"
    "log"
    "seu-projeto/initializers/gcs"
    "google.golang.org/api/iterator"
)

func listBuckets() {
    ctx := context.Background()
    credentialsPath := "/path/to/service-account.json"
    projectID := "meu-projeto-id"

    // Inicializar cliente
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    // Listar buckets
    buckets := gcsConfig.Buckets(ctx, projectID)
    for {
        bucketAttrs, err := buckets.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            log.Printf("Erro ao listar buckets: %v", err)
            break
        }
        fmt.Printf("Bucket: %s\n", bucketAttrs.Name)
    }
}
```

### Upload de Arquivo

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "seu-projeto/initializers/gcs"
)

func uploadFile() {
    ctx := context.Background()
    credentialsPath := "/path/to/service-account.json"
    bucketName := "meu-bucket"
    objectName := "uploads/exemplo.txt"

    // Inicializar cliente
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    // Obter referência do bucket e objeto
    bucket := gcsConfig.Bucket(bucketName)
    obj := bucket.Object(objectName)

    // Criar writer para upload
    w := obj.NewWriter(ctx)
    w.ContentType = "text/plain"
    w.Metadata = map[string]string{
        "uploaded-by": "go-application",
        "timestamp":   time.Now().Format(time.RFC3339),
    }

    // Escrever dados
    data := "Este é um exemplo de upload para Google Cloud Storage"
    if _, err := w.Write([]byte(data)); err != nil {
        log.Printf("Erro ao escrever dados: %v", err)
        return
    }

    // Finalizar upload
    if err := w.Close(); err != nil {
        log.Printf("Erro ao finalizar upload: %v", err)
        return
    }

    fmt.Printf("Arquivo '%s' enviado com sucesso para bucket '%s'\n", objectName, bucketName)
}
```

### Download de Arquivo

```go
package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "seu-projeto/initializers/gcs"
)

func downloadFile() {
    ctx := context.Background()
    credentialsPath := "/path/to/service-account.json"
    bucketName := "meu-bucket"
    objectName := "uploads/exemplo.txt"

    // Inicializar cliente
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    // Obter referência do objeto
    obj := gcsConfig.Bucket(bucketName).Object(objectName)

    // Criar reader para download
    r, err := obj.NewReader(ctx)
    if err != nil {
        log.Printf("Erro ao criar reader: %v", err)
        return
    }
    defer r.Close()

    // Ler dados
    data, err := io.ReadAll(r)
    if err != nil {
        log.Printf("Erro ao ler dados: %v", err)
        return
    }

    fmt.Printf("Conteúdo do arquivo '%s': %s\n", objectName, string(data))
    fmt.Printf("Tamanho: %d bytes\n", len(data))
    fmt.Printf("Content-Type: %s\n", r.ContentType())
}
```

### Listagem de Objetos

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "seu-projeto/initializers/gcs"
    "cloud.google.com/go/storage"
    "google.golang.org/api/iterator"
)

func listObjects() {
    ctx := context.Background()
    credentialsPath := "/path/to/service-account.json"
    bucketName := "meu-bucket"

    // Inicializar cliente
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    // Configurar query para listagem
    query := &storage.Query{
        Prefix:    "uploads/",
        Delimiter: "/",
    }

    // Listar objetos
    it := gcsConfig.Bucket(bucketName).Objects(ctx, query)
    for {
        attrs, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            log.Printf("Erro ao listar objetos: %v", err)
            break
        }

        fmt.Printf("Objeto: %s\n", attrs.Name)
        fmt.Printf("  Tamanho: %d bytes\n", attrs.Size)
        fmt.Printf("  Criado: %s\n", attrs.Created.Format(time.RFC3339))
        fmt.Printf("  Content-Type: %s\n", attrs.ContentType)
        if len(attrs.Metadata) > 0 {
            fmt.Printf("  Metadata: %+v\n", attrs.Metadata)
        }
        fmt.Println()
    }
}
```

### Exclusão de Objeto

```go
package main

import (
    "context"
    "fmt"
    "log"
    "seu-projeto/initializers/gcs"
)

func deleteObject() {
    ctx := context.Background()
    credentialsPath := "/path/to/service-account.json"
    bucketName := "meu-bucket"
    objectName := "uploads/exemplo.txt"

    // Inicializar cliente
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    // Deletar objeto
    obj := gcsConfig.Bucket(bucketName).Object(objectName)
    if err := obj.Delete(ctx); err != nil {
        log.Printf("Erro ao deletar objeto: %v", err)
        return
    }

    fmt.Printf("Objeto '%s' deletado com sucesso\n", objectName)
}
```

## Operações Avançadas

### URLs Assinadas

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "seu-projeto/initializers/gcs"
    "cloud.google.com/go/storage"
)

func generateSignedURL() {
    ctx := context.Background()
    credentialsPath := "/path/to/service-account.json"
    bucketName := "meu-bucket"
    objectName := "private/documento.pdf"

    // Inicializar cliente
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    // Configurar opções para URL assinada
    opts := &storage.SignedURLOptions{
        Scheme:  storage.SigningSchemeV4,
        Method:  "GET",
        Expires: time.Now().Add(1 * time.Hour),
    }

    // Gerar URL assinada
    url, err := gcsConfig.Bucket(bucketName).SignedURL(objectName, opts)
    if err != nil {
        log.Printf("Erro ao gerar URL assinada: %v", err)
        return
    }

    fmt.Printf("URL assinada para download: %s\n", url)
}
```

### Metadados de Objeto

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "seu-projeto/initializers/gcs"
)

func manageMetadata() {
    ctx := context.Background()
    credentialsPath := "/path/to/service-account.json"
    bucketName := "meu-bucket"
    objectName := "metadata-exemplo.txt"

    // Inicializar cliente
    gcsConfig := gcs.Initialize(ctx, credentialsPath)
    defer gcsConfig.Close()

    obj := gcsConfig.Bucket(bucketName).Object(objectName)

    // Upload com metadados customizados
    w := obj.NewWriter(ctx)
    w.ContentType = "text/plain"
    w.ContentLanguage = "pt-BR"
    w.CacheControl = "public, max-age=3600"
    w.Metadata = map[string]string{
        "author":      "Sistema Go",
        "version":     "1.0",
        "environment": "production",
        "created-at":  time.Now().Format(time.RFC3339),
    }

    data := "Arquivo com metadados customizados"
    if _, err := w.Write([]byte(data)); err != nil {
        log.Printf("Erro ao escrever: %v", err)
        return
    }

    if err := w.Close(); err != nil {
        log.Printf("Erro ao finalizar upload: %v", err)
        return
    }

    // Ler metadados
    attrs, err := obj.Attrs(ctx)
    if err != nil {
        log.Printf("Erro ao obter atributos: %v", err)
        return
    }

    fmt.Printf("Metadados do objeto '%s':\n", objectName)
    fmt.Printf("  Content-Type: %s\n", attrs.ContentType)
    fmt.Printf("  Content-Language: %s\n", attrs.ContentLanguage)
    fmt.Printf("  Cache-Control: %s\n", attrs.CacheControl)
    fmt.Printf("  Tamanho: %d bytes\n", attrs.Size)
    fmt.Printf("  Criado: %s\n", attrs.Created.Format(time.RFC3339))
    fmt.Printf("  Metadata customizada:\n")
    for key, value := range attrs.Metadata {
        fmt.Printf("    %s: %s\n", key, value)
    }
}
```

## Tratamento de Erros

### Captura de Erros de Inicialização

```go
package main

import (
    "context"
    "fmt"
    "log"
    "seu-projeto/initializers/gcs"
)

func handleInitializationErrors() {
    ctx := context.Background()

    // Exemplo de tratamento de erro com recover
    func() {
        defer func() {
            if r := recover(); r != nil {
                fmt.Printf("Erro capturado na inicialização: %v\n", r)
            }
        }()

        // Esta linha causará panic se as credenciais forem inválidas
        credentialsPath := "/path/to/invalid-credentials.json"
        gcsConfig := gcs.Initialize(ctx, credentialsPath)
        defer gcsConfig.Close()
    }()

    // Exemplo com credenciais válidas mas operação que pode falhar
    validCredentials := "/path/to/valid-credentials.json"
    gcsConfig := gcs.Initialize(ctx, validCredentials)
    defer gcsConfig.Close()

    // Tentar acessar bucket inexistente
    bucketName := "bucket-que-nao-existe-12345"
    obj := gcsConfig.Bucket(bucketName).Object("test.txt")

    _, err := obj.Attrs(ctx)
    if err != nil {
        fmt.Printf("Erro esperado ao acessar bucket inexistente: %v\n", err)
    }
}
```

## Configuração de Produção

### Variáveis de Ambiente

```bash
# Configurações do Google Cloud Storage
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
export GOOGLE_CLOUD_PROJECT="meu-projeto-id"
export GCS_BUCKET_NAME="meu-bucket-producao"
```

### Configuração via Arquivo

```yaml
# config/gcs.yaml
gcs:
  credentials_path: "/path/to/service-account.json"
  project_id: "meu-projeto-id"
  default_bucket: "meu-bucket"
  timeout: "30s"
  retry_attempts: 3
```

### Docker

```dockerfile
# Dockerfile
FROM golang:1.21-alpine

# Copiar credenciais do GCS
COPY service-account.json /app/credentials/

# Definir variável de ambiente
ENV GOOGLE_APPLICATION_CREDENTIALS=/app/credentials/service-account.json

# Copiar e compilar aplicação
COPY . /app
WORKDIR /app
RUN go build -o main .

CMD ["./main"]
```

## Melhores Práticas

### 1. Segurança

```go
// ✅ Usar Service Account com permissões mínimas
// ✅ Não hardcodar credenciais no código
// ✅ Usar variáveis de ambiente ou secret managers
// ✅ Configurar CORS adequadamente para aplicações web

// ❌ Evitar
// credentialsPath := "./my-secret-key.json" // hardcoded

// ✅ Recomendado
credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
if credentialsPath == "" {
    log.Fatal("Credenciais não configuradas")
}
```

### 2. Performance

```go
// ✅ Reutilizar cliente GCS
var gcsClient *gcs.GCSConfig

func init() {
    ctx := context.Background()
    credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
    gcsClient = gcs.Initialize(ctx, credentialsPath)
}

// ✅ Usar ChunkSize apropriado para uploads grandes
w := obj.NewWriter(ctx)
w.ChunkSize = 1024 * 1024 // 1MB chunks

// ✅ Configurar timeouts adequados
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 3. Organização

```go
// ✅ Usar prefixos consistentes
const (
    UploadsPrefix   = "uploads/"
    TempPrefix      = "temp/"
    ArchivePrefix   = "archive/"
)

// ✅ Estruturar nomes de objetos
objectName := fmt.Sprintf("%s%s/%s", UploadsPrefix, userID, filename)

// ✅ Usar metadados para classificação
w.Metadata = map[string]string{
    "user-id":     userID,
    "upload-type": "document",
    "version":     "1.0",
}
```

### 4. Monitoramento

```go
// ✅ Implementar logging estruturado
log.Printf("Upload iniciado: bucket=%s, object=%s, size=%d",
    bucketName, objectName, fileSize)

// ✅ Monitorar operações
start := time.Now()
defer func() {
    duration := time.Since(start)
    log.Printf("Upload concluído em %v", duration)
}()

// ✅ Implementar métricas
// Use bibliotecas como Prometheus para coletar métricas
```

## Dependências

- `cloud.google.com/go/storage` - Cliente oficial do Google Cloud Storage
- `google.golang.org/api/option` - Opções de configuração da API
- `google.golang.org/api/iterator` - Iteradores para listagens
- `context` - Gerenciamento de contexto

## Veja Também

- [Pacote Auth](../auth/README.md) - Para autenticação
- [Pacote Crypt](../crypt/README.md) - Para criptografia
- [Pacote OpenTelemetry](../opentelemetry/README.md) - Para observabilidade
- [Pacote Formatter](../formatter/README.md) - Para formatação de erros

---

**Nota**: Este pacote fornece apenas a função de inicialização do cliente GCS. Para operações específicas como upload, download, listagem, etc., utilize diretamente as funcionalidades do cliente `*storage.Client` retornado pela função `Initialize()`.
