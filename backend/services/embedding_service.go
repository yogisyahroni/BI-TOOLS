package services

import (
	"context"
	"errors"
	"fmt"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"insight-engine-backend/services/ai"
	"strings"
)

// EmbeddingService handles generation and storage of vector embeddings
type EmbeddingService struct {
	encryptionService *EncryptionService
	providerFactory   *ai.ProviderFactory
}

// NewEmbeddingService creates a new Embedding Service
func NewEmbeddingService(encryptionService *EncryptionService) *EmbeddingService {
	return &EmbeddingService{
		encryptionService: encryptionService,
		providerFactory:   ai.NewProviderFactory(),
	}
}

// getActiveProvider retrieves and decrypts the active provider for a user
func (s *EmbeddingService) getActiveProvider(userID string, providerID string) (ai.EmbeddingProvider, error) {
	var provider models.AIProvider
	var err error

	if providerID != "" {
		err = database.DB.Where("id = ? AND user_id = ?", providerID, userID).First(&provider).Error
	} else {
		err = database.DB.Where("user_id = ? AND is_default = ? AND is_active = ?", userID, true, true).First(&provider).Error
		if err != nil {
			err = database.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&provider).Error
		}
	}

	if err != nil {
		return nil, errors.New("provider not found or access denied")
	}

	if !provider.IsActive {
		return nil, errors.New("provider is not active")
	}

	apiKey, err := s.encryptionService.Decrypt(provider.APIKeyEncrypted)
	if err != nil {
		return nil, errors.New("failed to decrypt API key")
	}

	providerConfig := ai.ProviderConfig{
		Type:   provider.ProviderType,
		APIKey: apiKey,
		Model:  provider.Model,
		Config: provider.Config,
	}

	if provider.BaseURL != nil {
		providerConfig.BaseURL = *provider.BaseURL
	}

	aiProvider, err := s.providerFactory.CreateProvider(providerConfig)
	if err != nil {
		return nil, err
	}

	// Make sure the provider actually supports embeddings. We can do a type assertion here.
	embProvider, ok := aiProvider.(ai.EmbeddingProvider)
	if !ok {
		return nil, fmt.Errorf("provider %s does not support embeddings", provider.ProviderType)
	}

	return embProvider, nil
}

// GenerateEmbeddings simply generates embeddings for a slice of texts
func (s *EmbeddingService) GenerateEmbeddings(ctx context.Context, userID, providerID string, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	provider, err := s.getActiveProvider(userID, providerID)
	if err != nil {
		return nil, err
	}

	// We default to text-embedding-ada-002 if the user's primary model is a chat model, or text-embedding-3-small if available.
	// In a real app we'd have a specific "embedding_model" column.
	embeddingModel := "text-embedding-3-small"
	if provider.GetInfo().Type == "openai" {
		embeddingModel = "text-embedding-3-small"
	}

	req := ai.EmbeddingRequest{
		Input: texts,
		Model: embeddingModel,
	}

	resp, err := provider.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Embeddings, nil
}

// RetrieveRelevantSchema finds the most relevant tables for a given prompt using pgvector cosine similarity
func (s *EmbeddingService) RetrieveRelevantSchema(ctx context.Context, userID, providerID, connectionID, prompt string, limit int) ([]map[string]interface{}, error) {
	// 1. Generate embedding for the prompt
	embeddings, err := s.GenerateEmbeddings(ctx, userID, providerID, []string{prompt})
	if err != nil || len(embeddings) == 0 {
		return nil, err
	}

	promptEmbedding := embeddings[0]

	// Format the embedding as a pgvector string "[1.1, 2.2, ...]"
	var strEmbed strings.Builder
	strEmbed.WriteString("[")
	for j, val := range promptEmbedding {
		if j > 0 {
			strEmbed.WriteString(",")
		}
		strEmbed.WriteString(fmt.Sprintf("%f", val))
	}
	strEmbed.WriteString("]")

	// 2. Query the database using cosine similarity (<=>)
	// We join with the connections table to ensure the user owns the connection
	query := `
		SELECT se.schema_name, se.table_name, se.description, 1 - (se.embedding <=> ?::vector) as similarity
		FROM schema_embeddings se
		JOIN connections c ON se.connection_id = c.id
		WHERE se.connection_id = ? AND c.user_id = ?
		ORDER BY se.embedding <=> ?::vector
		LIMIT ?
	`

	rows, err := database.DB.Raw(query, strEmbed.String(), connectionID, userID, strEmbed.String(), limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var schemaName, tableName, description string
		var similarity float64

		if err := rows.Scan(&schemaName, &tableName, &description, &similarity); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"schema_name": schemaName,
			"table_name":  tableName,
			"description": description,
			"similarity":  similarity,
		})
	}

	return results, nil
}
