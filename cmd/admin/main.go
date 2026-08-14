package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/walfa-labs/backend/internal/config"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/platform"
	"github.com/walfa-labs/backend/internal/port"

	"github.com/walfa-labs/backend/internal/adapter/repository/oracle/atp"
	pgRepo "github.com/walfa-labs/backend/internal/adapter/repository/postgres"
)

const defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*-_=+"

func generateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 16
	}

	charsets := []string{
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"0123456789",
		"!@#$%^&*-_=+",
	}

	result := make([]byte, length)
	// Ensure at least one character from each required class
	for i, cs := range charsets {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(cs))))
		if err != nil {
			return "", err
		}
		result[i] = cs[idx.Int64()]
	}

	// Fill remainder from full charset
	for i := len(charsets); i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(defaultCharset))))
		if err != nil {
			return "", err
		}
		result[i] = defaultCharset[idx.Int64()]
	}

	// Fisher-Yates shuffle
	for i := len(result) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := int(jBig.Int64())
		result[i], result[j] = result[j], result[i]
	}

	return string(result), nil
}

func updateEnvFile(filePath, username, password, hash string) error {
	cleanPath := filepath.Clean(filePath)
	// #nosec G304 -- filePath is supplied via CLI flag by admin operator
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	userUpdated := false
	hashUpdated := false
	commentUpdated := false

	userRegex := regexp.MustCompile(`^\s*ADMIN_USERNAME=`)
	hashRegex := regexp.MustCompile(`^\s*ADMIN_PASSWORD_HASH=`)
	commentRegex := regexp.MustCompile(`^\s*#\s*Admin credentials\s*\(.*\)`)

	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if commentRegex.MatchString(trimmed) {
			newLines = append(newLines, fmt.Sprintf("# Admin credentials (%s / %s)", username, password))
			commentUpdated = true
		} else if userRegex.MatchString(trimmed) {
			newLines = append(newLines, fmt.Sprintf("ADMIN_USERNAME=%s", username))
			userUpdated = true
		} else if hashRegex.MatchString(trimmed) {
			newLines = append(newLines, fmt.Sprintf("ADMIN_PASSWORD_HASH='%s'", hash))
			hashUpdated = true
		} else {
			newLines = append(newLines, trimmed)
		}
	}

	if !commentUpdated && (!userUpdated || !hashUpdated) {
		newLines = append(newLines, fmt.Sprintf("# Admin credentials (%s / %s)", username, password))
	}
	if !userUpdated {
		newLines = append(newLines, fmt.Sprintf("ADMIN_USERNAME=%s", username))
	}
	if !hashUpdated {
		newLines = append(newLines, fmt.Sprintf("ADMIN_PASSWORD_HASH='%s'", hash))
	}

	output := strings.Join(newLines, "\n")
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	// #nosec G703 -- CLI tool writing to admin specified .env file
	return os.WriteFile(cleanPath, []byte(output), 0o600)
}

func loadDotEnv(filePath string) {
	cleanPath := filepath.Clean(filePath)
	// #nosec G304 -- filePath is supplied via CLI flag by admin operator
	file, err := os.Open(cleanPath)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			// Strip quotes if present
			if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
				v = v[1 : len(v)-1]
			}
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}

func main() {
	usernameFlag := flag.String("username", "", "Admin username (defaults to ADMIN_USERNAME env or 'admin')")
	passwordFlag := flag.String("password", "", "Admin cleartext password (if empty, generates secure random password)")
	lengthFlag := flag.Int("length", 16, "Generated password length")
	costFlag := flag.Int("cost", 10, "Bcrypt cost factor (4-31)")
	noDBFlag := flag.Bool("no-db", false, "Skip updating database")
	noEnvFlag := flag.Bool("no-env", false, "Skip updating .env file")
	envPathFlag := flag.String("env", ".env", "Path to .env file")
	flag.Parse()

	// Load .env first if present
	loadDotEnv(*envPathFlag)

	username := *usernameFlag
	if username == "" {
		username = os.Getenv("ADMIN_USERNAME")
		if username == "" {
			username = "admin"
		}
	}

	password := *passwordFlag
	if password == "" {
		gen, err := generateRandomPassword(*lengthFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating password: %v\n", err)
			os.Exit(1)
		}
		password = gen
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), *costFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating bcrypt hash: %v\n", err)
		os.Exit(1)
	}
	hash := string(hashBytes)

	// Update .env if requested
	envUpdated := false
	if !*noEnvFlag {
		if err := updateEnvFile(*envPathFlag, username, password, hash); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update %s: %v\n", *envPathFlag, err)
		} else {
			envUpdated = true
		}
	}

	// Update Database if requested
	dbUpdated := false
	var dbDriver string
	if !*noDBFlag {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load database config: %v (skipping DB update)\n", err)
		} else {
			dbDriver = cfg.DBDriver
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var adminRepo port.AdminRepo
			switch cfg.DBDriver {
			case "postgres":
				pgDB, err := platform.NewPostgresDB(ctx, cfg.PostgresDSN())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to connect to PostgreSQL: %v\n", err)
				} else {
					defer func() { _ = pgDB.Close() }()
					adminRepo = pgRepo.NewAdminRepo(pgDB)
				}
			default:
				atpDB, err := platform.NewOracleDB(ctx, cfg.OracleDSN())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to connect to Oracle: %v\n", err)
				} else {
					defer func() { _ = atpDB.Close() }()
					adminRepo = atp.NewAdminRepo(atpDB)
				}
			}

			if adminRepo != nil {
				// Check if user already exists to reuse ID
				var userID uuid.UUID
				existing, err := adminRepo.GetByUsername(ctx, username)
				if err == nil && existing != nil {
					userID = existing.ID
				} else {
					userID = uuid.New()
				}

				adminUser := &domain.AdminUser{
					ID:           userID,
					Username:     username,
					PasswordHash: hash,
					CreatedAt:    time.Now(),
				}

				if err := adminRepo.Upsert(ctx, adminUser); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to upsert admin user to database: %v\n", err)
				} else {
					dbUpdated = true
				}
			}
		}
	}

	// Output summary
	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("       Admin Credentials Regenerated")
	fmt.Println("==================================================")
	fmt.Printf(" Username : %s\n", username)
	fmt.Printf(" Password : %s\n", password)
	fmt.Printf(" Hash     : %s\n", hash)
	fmt.Println("--------------------------------------------------")
	if dbUpdated {
		fmt.Printf(" Database : Updated successfully (%s)\n", dbDriver)
	} else if *noDBFlag {
		fmt.Println(" Database : Skipped (-no-db)")
	} else {
		fmt.Println(" Database : Not updated (connection unavailable)")
	}

	if envUpdated {
		fmt.Printf(" .env     : Updated successfully (%s)\n", *envPathFlag)
	} else if *noEnvFlag {
		fmt.Println(" .env     : Skipped (-no-env)")
	} else {
		fmt.Println(" .env     : Not updated")
	}
	fmt.Println("==================================================")
	fmt.Println(" Test admin login:")
	fmt.Printf(" curl -s -X POST http://localhost:8080/api/v1/auth/login \\\n")
	fmt.Printf("   -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("   -d '{\"username\":\"%s\",\"password\":\"%s\"}'\n", username, password)
	fmt.Println("==================================================")
	fmt.Println()
}
