package platform

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

// FiberConfig returns a fiber.Config pre-wired with Sonic as the JSON
// encoder/decoder. Sonic is a JIT-based, allocation-free JSON library that
// outperforms encoding/json on amd64/arm64.
func FiberConfig() fiber.Config {
	return fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
		AppName:     "Portfolio API v1",
	}
}
