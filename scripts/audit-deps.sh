#!/bin/bash
# Audit Dependencies - Verifica violações arquiteturais
echo "🔍 Auditoria de Dependências"
echo ""
cd "$(dirname "$0")/.."

VIOLATIONS=0

echo "1. Verificando config/adapter..."
if [ -d "config/adapter" ]; then
    echo "  ❌ VIOLAÇÃO: config/adapter/ existe"
    echo "     Isso cria acoplamento entre config e pkg"
    VIOLATIONS=$((VIOLATIONS + 1))
else
    echo "  ✅ OK"
fi

echo ""
echo "2. Verificando imports de config em pkg..."
for pkg in pkg/cache pkg/logger; do
    imports=$(find "$pkg" -name "*.go" -not -name "*_test.go" -exec grep -l "github.com/marcelofabianov/config" {} \; 2>/dev/null || true)
    if [ -n "$imports" ]; then
        echo "  ❌ VIOLAÇÃO: $pkg importa config"
        VIOLATIONS=$((VIOLATIONS + 1))
    else
        echo "  ✅ $pkg: OK"
    fi
done

echo ""
if [ $VIOLATIONS -eq 0 ]; then
    echo "✅ Arquitetura limpa!"
    exit 0
else
    echo "❌ $VIOLATIONS violações encontradas"
    exit 1
fi
