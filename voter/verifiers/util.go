package verifiers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	solana "github.com/olegfomenko/solana-go"
)

type Metadata struct {
	Image string `json:"image"`
}

func GetImage(metaUrl string) (url string, hash string, err error) {
	metaResp, err := http.Get(metaUrl)
	if err != nil {
		return "", "", err
	}

	defer metaResp.Body.Close()
	b, err := io.ReadAll(metaResp.Body)
	if err != nil {
		return "", "", err
	}

	var meta = new(Metadata)
	if err := json.Unmarshal(b, meta); err != nil {
		return "", "", err
	}

	imageResp, err := http.Get(meta.Image)
	if err != nil {
		return "", "", err
	}

	defer imageResp.Body.Close()
	b, err = io.ReadAll(imageResp.Body)
	if err != nil {
		return "", "", err
	}

	imgHash := sha256.Sum256(b)
	hash = base64.StdEncoding.EncodeToString(imgHash[:])
	return meta.Image, hash, nil
}

func GenerateTokenSeed(bridgeContract string) (seedStr string, idStr string) {
	programId := solana.PublicKeyFromBytes(hexutil.MustDecode(bridgeContract))

	for {
		var seed [32]byte
		_, err := rand.Read(seed[:])
		if err != nil {
			panic(err)
		}

		key, _, err := solana.FindProgramAddress([][]byte{seed[:]}, programId)
		if err == nil {
			seedStr = hexutil.Encode(seed[:])
			idStr = hexutil.Encode(key.Bytes())
			return
		}
	}
}

func VerifyTokenSeed(bridgeContract, tokenSeed string) bool {
	programId := solana.PublicKeyFromBytes(hexutil.MustDecode(bridgeContract))
	seed, err := hexutil.Decode(tokenSeed)
	if err != nil {
		return false
	}

	_, _, err = solana.FindProgramAddress([][]byte{seed}, programId)
	return err == nil
}
