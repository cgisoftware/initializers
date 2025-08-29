package gcs

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// ExampleBasicInitialization demonstra inicialização básica do cliente GCS
func ExampleBasicInitialization() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"

	// Inicializar cliente GCS usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
	defer gcsConfig.Close()

	fmt.Println("Cliente GCS inicializado com sucesso")

	// Exemplo de uso básico - listar buckets
	buckets := gcsConfig.Buckets(ctx, "my-project-id")
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

// ExampleUploadFile demonstra como fazer upload de arquivos
func ExampleUploadFile() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"
	bucketName := "my-bucket"
	objectName := "uploads/example.txt"

	// Inicializar cliente usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
	defer gcsConfig.Close()

	// Obter referência do bucket
	bucket := gcsConfig.Bucket(bucketName)

	// Criar objeto
	obj := bucket.Object(objectName)

	// Criar writer
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

	// Fechar writer (finaliza upload)
	if err := w.Close(); err != nil {
		log.Printf("Erro ao finalizar upload: %v", err)
		return
	}

	fmt.Printf("Arquivo '%s' enviado com sucesso para bucket '%s'\n", objectName, bucketName)
}

// ExampleDownloadFile demonstra como fazer download de arquivos
func ExampleDownloadFile() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"
	bucketName := "my-bucket"
	objectName := "uploads/example.txt"

	// Inicializar cliente usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
	defer gcsConfig.Close()

	// Obter referência do objeto
	obj := gcsConfig.Bucket(bucketName).Object(objectName)

	// Criar reader
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

// ExampleListObjects demonstra como listar objetos em um bucket
func ExampleListObjects() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"
	bucketName := "my-bucket"

	// Inicializar cliente usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
	defer gcsConfig.Close()

	// Listar objetos com prefixo
	query := &storage.Query{
		Prefix:    "uploads/",
		Delimiter: "/",
	}

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

// ExampleDeleteObject demonstra como deletar objetos
func ExampleDeleteObject() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"
	bucketName := "my-bucket"
	objectName := "uploads/example.txt"

	// Inicializar cliente usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
	defer gcsConfig.Close()

	// Deletar objeto
	obj := gcsConfig.Bucket(bucketName).Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		log.Printf("Erro ao deletar objeto: %v", err)
		return
	}

	fmt.Printf("Objeto '%s' deletado com sucesso\n", objectName)
}

// ExampleObjectMetadata demonstra como trabalhar com metadados
func ExampleObjectMetadata() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"
	bucketName := "my-bucket"
	objectName := "metadata-example.txt"

	// Inicializar cliente usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
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
	fmt.Printf("  MD5: %x\n", attrs.MD5)
	fmt.Printf("  Criado: %s\n", attrs.Created.Format(time.RFC3339))
	fmt.Printf("  Atualizado: %s\n", attrs.Updated.Format(time.RFC3339))
	fmt.Printf("  Metadata customizada:\n")
	for key, value := range attrs.Metadata {
		fmt.Printf("    %s: %s\n", key, value)
	}
}

// ExampleSignedURL demonstra como gerar URLs assinadas
func ExampleSignedURL() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"
	bucketName := "my-bucket"
	objectName := "private/document.pdf"

	// Inicializar cliente usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
	defer gcsConfig.Close()

	// Gerar URL assinada para download (válida por 1 hora)
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(1 * time.Hour),
	}

	url, err := gcsConfig.Bucket(bucketName).SignedURL(objectName, opts)
	if err != nil {
		log.Printf("Erro ao gerar URL assinada: %v", err)
		return
	}

	fmt.Printf("URL assinada para download: %s\n", url)

	// Gerar URL assinada para upload
	uploadOpts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     time.Now().Add(30 * time.Minute),
		ContentType: "application/pdf",
	}

	uploadURL, err := gcsConfig.Bucket(bucketName).SignedURL("uploads/new-document.pdf", uploadOpts)
	if err != nil {
		log.Printf("Erro ao gerar URL de upload: %v", err)
		return
	}

	fmt.Printf("URL assinada para upload: %s\n", uploadURL)
}

// ExampleBatchOperations demonstra operações em lote
func ExampleBatchOperations() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"
	bucketName := "my-bucket"

	// Inicializar cliente usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
	defer gcsConfig.Close()

	bucket := gcsConfig.Bucket(bucketName)

	// Upload múltiplos arquivos
	files := []struct {
		name    string
		content string
	}{
		{"batch/file1.txt", "Conteúdo do arquivo 1"},
		{"batch/file2.txt", "Conteúdo do arquivo 2"},
		{"batch/file3.txt", "Conteúdo do arquivo 3"},
	}

	fmt.Println("Iniciando upload em lote...")
	for _, file := range files {
		obj := bucket.Object(file.name)
		w := obj.NewWriter(ctx)
		w.ContentType = "text/plain"

		if _, err := w.Write([]byte(file.content)); err != nil {
			log.Printf("Erro ao escrever %s: %v", file.name, err)
			continue
		}

		if err := w.Close(); err != nil {
			log.Printf("Erro ao finalizar upload %s: %v", file.name, err)
			continue
		}

		fmt.Printf("✓ Upload concluído: %s\n", file.name)
	}

	// Listar arquivos criados
	fmt.Println("\nArquivos criados:")
	query := &storage.Query{Prefix: "batch/"}
	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Erro ao listar: %v", err)
			break
		}
		fmt.Printf("  %s (%d bytes)\n", attrs.Name, attrs.Size)
	}
}

// ExampleStreamingUpload demonstra upload de stream grande
func ExampleStreamingUpload() {
	ctx := context.Background()
	credentialsPath := "/path/to/service-account.json"
	bucketName := "my-bucket"
	objectName := "streaming/large-file.txt"

	// Inicializar cliente usando o pacote
	gcsConfig := Initialize(ctx, credentialsPath)
	defer gcsConfig.Close()

	obj := gcsConfig.Bucket(bucketName).Object(objectName)
	w := obj.NewWriter(ctx)
	w.ContentType = "text/plain"
	w.ChunkSize = 1024 * 1024 // 1MB chunks

	// Simular stream de dados grande
	fmt.Println("Iniciando upload streaming...")
	for i := 0; i < 100; i++ {
		data := fmt.Sprintf("Chunk %d: %s\n", i, strings.Repeat("data", 100))
		if _, err := w.Write([]byte(data)); err != nil {
			log.Printf("Erro ao escrever chunk %d: %v", i, err)
			break
		}

		if i%10 == 0 {
			fmt.Printf("Progresso: %d%%\n", i)
		}
	}

	if err := w.Close(); err != nil {
		log.Printf("Erro ao finalizar upload: %v", err)
		return
	}

	fmt.Println("Upload streaming concluído!")
}

// ExampleErrorHandling demonstra tratamento de erros
func ExampleErrorHandling() {
	ctx := context.Background()

	// Tentar inicializar com credenciais inválidas
	fmt.Println("=== TRATAMENTO DE ERROS ===")

	// Este exemplo mostra como capturar erros de inicialização
	// Em produção, use defer recover() se necessário
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Erro capturado na inicialização: %v\n", r)
			}
		}()

		// Esta linha causará panic se as credenciais forem inválidas
		// credentialsPath := "/path/to/invalid-credentials.json"
		// gcsConfig := Initialize(ctx, credentialsPath)
		// defer gcsConfig.Close()
	}()

	// Exemplo com credenciais válidas mas bucket inexistente
	validCredentials := "/path/to/valid-credentials.json"
	gcsConfig := Initialize(ctx, validCredentials)
	defer gcsConfig.Close()

	// Tentar acessar bucket inexistente
	bucketName := "bucket-que-nao-existe-12345"
	obj := gcsConfig.Bucket(bucketName).Object("test.txt")

	_, err := obj.Attrs(ctx)
	if err != nil {
		fmt.Printf("Erro esperado ao acessar bucket inexistente: %v\n", err)
	}

	// Tentar fazer upload para bucket sem permissão
	bucketSemPermissao := "bucket-sem-permissao"
	obj2 := gcsConfig.Bucket(bucketSemPermissao).Object("test.txt")
	w := obj2.NewWriter(ctx)

	if _, err := w.Write([]byte("test")); err != nil {
		fmt.Printf("Erro de permissão: %v\n", err)
	}

	if err := w.Close(); err != nil {
		fmt.Printf("Erro ao finalizar upload sem permissão: %v\n", err)
	}
}

// ExampleBestPractices demonstra melhores práticas
func ExampleBestPractices() {
	fmt.Println("=== MELHORES PRÁTICAS PARA GCS ===")
	fmt.Println("")
	fmt.Println("1. AUTENTICAÇÃO:")
	fmt.Println("   - Use Service Account com permissões mínimas necessárias")
	fmt.Println("   - Armazene credenciais de forma segura (não no código)")
	fmt.Println("   - Use variáveis de ambiente ou secret managers")
	fmt.Println("")
	fmt.Println("2. PERFORMANCE:")
	fmt.Println("   - Use ChunkSize apropriado para uploads grandes")
	fmt.Println("   - Implemente retry logic para operações")
	fmt.Println("   - Use conexões persistentes (reutilize cliente)")
	fmt.Println("   - Configure timeouts adequados")
	fmt.Println("")
	fmt.Println("3. SEGURANÇA:")
	fmt.Println("   - Use URLs assinadas para acesso temporário")
	fmt.Println("   - Configure CORS adequadamente")
	fmt.Println("   - Use criptografia em trânsito e em repouso")
	fmt.Println("   - Monitore acessos e operações")
	fmt.Println("")
	fmt.Println("4. ORGANIZAÇÃO:")
	fmt.Println("   - Use prefixos consistentes para organizar objetos")
	fmt.Println("   - Implemente lifecycle policies")
	fmt.Println("   - Use metadados para classificação")
	fmt.Println("   - Configure versionamento quando necessário")
	fmt.Println("")
	fmt.Println("5. MONITORAMENTO:")
	fmt.Println("   - Configure logging de operações")
	fmt.Println("   - Monitore custos e uso")
	fmt.Println("   - Implemente alertas para falhas")
	fmt.Println("   - Use métricas do Cloud Monitoring")
}
