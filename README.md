### Инициализация микросервисов
Каждый микросервис в проекте является отдельным модулем Go
1. Инициализация logger
```bash
   cd pkg/logger
   go mod init github.com/verbitsky-vladislav/blockchains-observer/pkg/logger
   go mod tidy
```
2. Инициализация eth-parser
```bash
   cd services/parsers/eth-parser
   go mod init github.com/verbitsky-vladislav/blockchains-observer/services/parsers/eth-parser
   go mod tidy
```
3. Инициализация btc-parser
```bash
   cd services/parsers/btc-parser
   go mod init github.com/verbitsky-vladislav/blockchains-observer/services/parsers/btc-parser
   go mod tidy
```