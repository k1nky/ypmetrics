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
	"google.golang.org/grpc/metadata"
)

// seal генератор подписи запросов.
type seal struct {
	hashers sync.Pool
}

// newSeal возвращает генератор подписей с ключом secret.
func newSeal(secret string) *seal {
	return &seal{
		hashers: sync.Pool{
			New: func() any {
				return hmac.New(sha256.New, []byte(secret))
			},
		},
	}
}

// SealUnaryInterceptor прерыватель для унарных запросов для формирования подписи передаваемых данных.
// Полученная подпись добавляется в meta-данные с ключом hashsha256.
func SealUnaryInterceptor(secret string) grpc.UnaryClientInterceptor {
	seal := newSeal(secret)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		h, err := seal.Use(req)
		if err != nil {
			return err
		}
		newCtx := metadata.AppendToOutgoingContext(ctx, "hashsha256", h)
		return invoker(newCtx, method, req, reply, cc, opts...)
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
