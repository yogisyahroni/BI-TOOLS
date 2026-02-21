package handlers

import (
	"insight-engine-backend/database"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// Global AI handlers (initialized in main.go)
var (
	aiProviderHandler *AIProviderHandler
	aiHandler         *AIHandler
)

// InitAIHandlers initializes AI handlers with encryption service
func InitAIHandlers(encryptionService *services.EncryptionService) {
	// Initialize Embedding Service (for RAG)
	embeddingService := services.NewEmbeddingService(encryptionService)
	aiProviderHandler = NewAIProviderHandler(encryptionService)

	// Initialize Semantic Layer Service
	semanticLayerService := services.NewSemanticLayerService(database.DB)

	// 2.7. Initialize Semantic Handlers
	// Initialize AI Service
	aiService := services.NewAIService(encryptionService, embeddingService, semanticLayerService)
	aiReasoningService := services.NewAIReasoningService(aiService)
	aiOptimizerService := services.NewAIOptimizerService(aiService)
	storyGeneratorService := services.NewStoryGeneratorService(aiService)
	aiHandler = NewAIHandler(aiService, aiReasoningService, aiOptimizerService, storyGeneratorService)
	InitSemanticHandlers(aiService)
}

// AI Provider Handler Wrappers
var GetAIProviders = func(c *fiber.Ctx) error {
	return aiProviderHandler.GetProviders(c)
}

var CreateAIProvider = func(c *fiber.Ctx) error {
	return aiProviderHandler.CreateProvider(c)
}

var GetAIProvider = func(c *fiber.Ctx) error {
	return aiProviderHandler.GetProvider(c)
}

var UpdateAIProvider = func(c *fiber.Ctx) error {
	return aiProviderHandler.UpdateProvider(c)
}

var DeleteAIProvider = func(c *fiber.Ctx) error {
	return aiProviderHandler.DeleteProvider(c)
}

var TestAIProvider = func(c *fiber.Ctx) error {
	return aiProviderHandler.TestProvider(c)
}

// AI Handler Wrappers
var GenerateAI = func(c *fiber.Ctx) error {
	return aiHandler.Generate(c)
}

var GetAIRequests = func(c *fiber.Ctx) error {
	return aiHandler.GetRequests(c)
}

var GetAIRequest = func(c *fiber.Ctx) error {
	return aiHandler.GetRequest(c)
}

var GetAIStats = func(c *fiber.Ctx) error {
	return aiHandler.GetUsageStats(c)
}

var StreamGenerateAI = func(c *fiber.Ctx) error {
	return aiHandler.StreamGenerate(c)
}

var ReasonAI = func(c *fiber.Ctx) error {
	return aiHandler.Reason(c)
}

var OptimizeAI = func(c *fiber.Ctx) error {
	return aiHandler.Optimize(c)
}
