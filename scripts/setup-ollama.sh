#!/bin/bash

# Setup script for Ollama models
# This script pulls the required models for Hermit

set -e

echo "🦙 Setting up Ollama models for Hermit..."
echo ""

# Check if Ollama is running
if ! curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "❌ Error: Ollama is not running on localhost:11434"
    echo "Please start Ollama first:"
    echo "  - If using docker-compose: docker-compose up -d ollama"
    echo "  - If installed locally: ollama serve"
    exit 1
fi

echo "✅ Ollama is running"
echo ""

# Pull embedding model
EMBED_MODEL="${OLLAMA_MODEL:-mxbai-embed-large}"
echo "📥 Pulling embedding model: $EMBED_MODEL"
echo "This may take a few minutes (~500MB download)..."
ollama pull "$EMBED_MODEL"
echo "✅ Embedding model ready"
echo ""

# Pull LLM model
LLM_MODEL="${OLLAMA_LLM_MODEL:-llama3.1}"
echo "📥 Pulling LLM model: $LLM_MODEL"
echo "This may take several minutes (~4.7GB download)..."
ollama pull "$LLM_MODEL"
echo "✅ LLM model ready"
echo ""

# Verify models
echo "🔍 Verifying installed models..."
ollama list

echo ""
echo "🎉 Setup complete! Hermit is ready to use."
echo ""
echo "Available models:"
echo "  - Embedding: $EMBED_MODEL (1024 dimensions)"
echo "  - LLM: $LLM_MODEL"
echo ""
echo "You can now start the Hermit API server."
