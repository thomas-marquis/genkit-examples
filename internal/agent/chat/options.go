package chat

type Option func(*Flow)

func WithLLM(modelName string) Option {
	return func(f *Flow) {
		f.llmModelName = modelName
	}
}

func WithEmbedding(modelName string) Option {
	return func(f *Flow) {
		f.embeddingModelName = modelName
	}
}

func WithNotionApiKey(apiKey string) Option {
	return func(f *Flow) {
		f.notionApiKey = apiKey
	}
}
