package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"hash"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrIncorrectSeal = status.Error(codes.Aborted, "request seal is incorrect")
)

// seal это middleware для проверки запроса.
// Подпись будет проставляться в MetaData HashSHA256.
type seal struct {
	hashers sync.Pool
}

// newSeal возвращает новую middleware для подписи с ключом secret.
func newSeal(secret string) *seal {
	return &seal{
		hashers: sync.Pool{
			New: func() any {
				return hmac.New(sha256.New, []byte(secret))
			},
		},
	}
}

func SealUnaryInterceptor(secret string) grpc.UnaryServerInterceptor {
	seal := newSeal(secret)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		h, err := seal.Use(req)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrIncorrectSeal
		}
		header := md.Get("hashsha256")
		if len(header) == 0 {
			return nil, ErrIncorrectSeal
		}
		if header[0] != h {
			return nil, ErrIncorrectSeal
		}
		return handler(ctx, req)
	}
}

// Use добавляет заголовок HashSHA256 с подписью передаваемых данных по алгоритму sha256.
func (s *seal) Use(m any) (string, error) {
	h := s.hashers.Get().(hash.Hash)
	defer s.hashers.Put(h)
	h.Reset()

	b := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(b).Encode(m); err != nil {
		return "", err
	}
	if _, err := h.Write(b.Bytes()); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
