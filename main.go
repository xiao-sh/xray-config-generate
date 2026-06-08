package main

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdp/qrterminal/v3"
)

type VlessConfig struct {
	ID          string
	Server      string
	Port        int
	Security    string
	Encryption  string
	PublicKey   string
	Fingerprint string
	Type        string
	SNI         string
	ShortID     string
	Remark      string
}
type Config struct {
	Log       Log        `json:"log"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
}

type Log struct {
	Loglevel string `json:"loglevel"`
}

type Inbound struct {
	Listen         string          `json:"listen"`
	Port           int             `json:"port"`
	Protocol       string          `json:"protocol"`
	Settings       InboundSettings `json:"settings"`
	StreamSettings StreamSettings  `json:"streamSettings"`
}

type InboundSettings struct {
	Clients    []Client `json:"clients"`
	Decryption string   `json:"decryption"`
}

type Client struct {
	ID string `json:"id"`
}

type StreamSettings struct {
	Network         string          `json:"network"`
	Security        string          `json:"security"`
	RealitySettings RealitySettings `json:"realitySettings"`
}

type RealitySettings struct {
	Show        bool     `json:"show"`
	Dest        string   `json:"dest"`
	Xver        int      `json:"xver"`
	ServerNames []string `json:"serverNames"`
	PrivateKey  string   `json:"privateKey"`
	ShortIds    []string `json:"shortIds"`
}

type Outbound struct {
	Protocol string `json:"protocol"`
}

func (v *VlessConfig) ShareLink() string {
	params := url.Values{}

	params.Set("security", v.Security)
	params.Set("encryption", v.Encryption)
	params.Set("pbk", v.PublicKey)
	params.Set("fp", v.Fingerprint)
	params.Set("type", v.Type)
	params.Set("sni", v.SNI)
	params.Set("sid", v.ShortID)

	return fmt.Sprintf(
		"vless://%s@%s:%d?%s#%s",
		v.ID,
		v.Server,
		v.Port,
		params.Encode(),
		url.QueryEscape(v.Remark),
	)
}

func GenerateShortID(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func GetPublicIP() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// resp, err := client.Get("https://api.ipify.org")
	resp, err := client.Get("https://ip.3322.net")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func main() {
	var port int
	var id string
	var shortID string
	var privateBytesString string
	var publicBytesString string
	var ip string

	id = uuid.New().String()

	curve := ecdh.X25519()

	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	privateBytes := privateKey.Bytes()
	publicBytes := privateKey.PublicKey().Bytes()

	privateBytesString = base64.RawURLEncoding.EncodeToString(privateBytes)
	publicBytesString = base64.RawURLEncoding.EncodeToString(publicBytes)

	// fmt.Println("Private:", base64.RawURLEncoding.EncodeToString(privateBytes))
	// fmt.Println("privateBytesString=", privateBytesString)

	// fmt.Println("Public:", base64.RawURLEncoding.EncodeToString(publicBytes))
	// fmt.Println("publicBytesString=", publicBytesString)

	ip, err = GetPublicIP()
	if err != nil {
		fmt.Println("获取公网IP失败:", err)
		return
	}

	// fmt.Println("生成的 UUID:", id)
	shortID, err = GenerateShortID(8)
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成 shortID 失败: %v\n", err)
		os.Exit(1)
	}
	// fmt.Println("生成的 ShortID:", shortID)
	flag.IntVar(&port, "port", 2025, "xray listen port")
	flag.Parse()
	cfg := Config{
		Log: Log{Loglevel: "warning"},
		Inbounds: []Inbound{
			{
				Listen:   "0.0.0.0",
				Port:     port,
				Protocol: "vless",
				Settings: InboundSettings{
					Clients:    []Client{{ID: id}},
					Decryption: "none",
				},
				StreamSettings: StreamSettings{
					Network:  "tcp",
					Security: "reality",
					RealitySettings: RealitySettings{
						Show:        false,
						Dest:        "www.yahoo.com:443",
						Xver:        0,
						ServerNames: []string{"www.yahoo.com"},
						PrivateKey:  privateBytesString,
						ShortIds:    []string{shortID},
					},
				},
			},
		},
		Outbounds: []Outbound{{Protocol: "freedom"}},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal config failed: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("config_simple.json", append(data, '\n'), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write file failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("config_simple.json 已生成")

	clientCfg := VlessConfig{
		ID:          id,
		Server:      ip,
		Port:        2025,
		Security:    "reality",
		Encryption:  "none",
		PublicKey:   publicBytesString,
		Fingerprint: "chrome",
		Type:        "tcp",
		SNI:         "www.yahoo.com",
		ShortID:     shortID,
		Remark:      "vless with reality",
	}

	fmt.Println(clientCfg.ShareLink())

	fmt.Println("\n二维码:")
	qrterminal.GenerateHalfBlock(
		clientCfg.ShareLink(),
		qrterminal.L,
		os.Stdout,
	)
}
