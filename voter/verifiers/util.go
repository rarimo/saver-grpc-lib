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

// MustGenerateTokenSeed returns seed and id hex-encoded with leading 0x
func MustGenerateTokenSeed(bridgeContract string) (string, string) {
	programId := solana.PublicKeyFromBytes(hexutil.MustDecode(bridgeContract))

	for {
		var seed [32]byte
		_, err := rand.Read(seed[:])
		if err != nil {
			panic(err)
		}

		key, _, err := solana.FindProgramAddress([][]byte{seed[:]}, programId)
		if err != nil {
			continue
		}

		return hexutil.Encode(seed[:]), hexutil.Encode(key.Bytes())
	}
}

func MustVerifyTokenSeed(bridgeContract, tokenSeed string) bool {
	seed, err := hexutil.Decode(tokenSeed)
	if err != nil {
		return false
	}

	_, _, err = solana.FindProgramAddress([][]byte{seed}, MustPublicKeyFromHexStr(bridgeContract))
	return err == nil
}

func MustGetPDA(bridgeContract, tokenSeed string) string {
	seed, err := hexutil.Decode(tokenSeed)
	if err != nil {
		return ""
	}

	key, _, err := solana.FindProgramAddress([][]byte{seed}, MustPublicKeyFromHexStr(bridgeContract))
	if err != nil {
		return ""
	}

	return hexutil.Encode(key.Bytes())
}

func MustPublicKeyFromHexStr(programId string) solana.PublicKey {
	return solana.PublicKeyFromBytes(hexutil.MustDecode(programId))
}
