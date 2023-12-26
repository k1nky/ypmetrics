package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func random(n int) ([]byte, error) {
	data := make([]byte, n)
	_, err := rand.Read(data)
	return data, err
}

func Test_EncrtyptRSA(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	text := []byte("abcdef12345")
	encrypted, err := EncryptRSA(&key.PublicKey, text)
	assert.NoError(t, err)
	assert.Equal(t, key.Size(), len(encrypted))
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, encrypted, nil)
	assert.NoError(t, err)
	assert.Equal(t, text, decrypted)
}

func Test_DecrtyptRSA(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	text := []byte("abcdef12345")
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, text, nil)
	assert.NoError(t, err)
	decrypted, err := DecryptRSA(key, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, text, decrypted)
}

func Test_EncryptDecrypt(t *testing.T) {
	type args struct {
		count int
		bs    int
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "100", args: args{count: 1, bs: 100}},
		{name: "1000", args: args{count: 10, bs: 100}},
		{name: "10000", args: args{count: 100, bs: 100}},
	}

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error(err)
		return
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := random(tt.args.bs)
			if err != nil {
				t.Error(err)
			}
			text := bytes.Repeat(b, tt.args.count)
			encrypted, err := EncryptRSA(&key.PublicKey, text)
			if err != nil {
				t.Error(err)
				return
			}
			decrypted, err := DecryptRSA(key, encrypted)
			if err != nil {
				t.Error(err)
				return
			}
			assert.Equal(t, text, decrypted, "")
		})
	}

}

func Test_chunkBytes(t *testing.T) {
	type args struct {
		src       []byte
		chunkSize int
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			name: "Chunk size less than size of slice #1",
			args: args{src: []byte{1, 2, 3, 4, 5, 6, 7}, chunkSize: 3},
			want: [][]byte{
				{1, 2, 3}, {4, 5, 6}, {7},
			},
		},
		{
			name: "Chunk size less than size of slice #2",
			args: args{src: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}, chunkSize: 3},
			want: [][]byte{
				{1, 2, 3}, {4, 5, 6}, {7, 8, 9},
			},
		},
		{
			name: "Chunk size greater than size of slice #1",
			args: args{src: []byte{1, 2}, chunkSize: 3},
			want: [][]byte{
				{1, 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chunkBytes(tt.args.src, tt.args.chunkSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("chunkBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ReadPrivateKey(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid format",
			args: args{strings.NewReader(`
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDXBwaV1+sK0IQ8
TaXxOmczKFrutKfN+WjgTPQf599+GZNCHEcb/2mnsNiFrqswaNPNnSwI34Q1uYlE
vUEZ40VdlHmqzd7PVfPLAO/7oeFqjWcKhEHXdnTVy9pnETuhlwPnX6QWtvCgRNJ2
2KuK1x36gGK1UBQAlxuxL8kmb/iAuIwkd+RIfbKX7p7C7/bOgQVBJT85HYUw0u90
5cCJLArzXljo3MS+akq6FRXyEtH4daIA96Qg13v+ehM3gpd8Eo/EWWTrXLbbyP51
GrMmGuqfl+nhoQFkNYKU6afdEtiQozGHCqsnISfnig2BoYvs45H6s4DoxygvZVAV
rajIp7LTAgMBAAECggEBAM01F0PJW7ZmaMxkDgm5AuP/j0hfgAVCEKR+zDvmvZNK
NQ7SjcpaZipyyzBJEXaRlBCV/oA5T6M1/ZpsQsTB8GDuYW5wKkMUdCU4L86lHsrh
R4Tx0yQdGEMw2K2j9JSx2jflPmOvEtTg2ToybQODaEi4XXtLgMtPak4enIjiMvYj
aRk7QbtM6F24w3ykTx7w/fyT6CQTKJTlBEGp9sMP6SqbtQcoYOqwZlhC8mW2TaU2
smRnHjOTCZTYvRn9iKYplvch/B7ZhIRdR39KiRW+4ghDHuGhVoINcbQhe0L01gTg
eKq30fQ4KmPKB0dNqqgVpJUbJQlLGdDfcBI+zXx1WAECgYEA8YZbch3JCKkbrxJ6
bEovzqRfdWm50wnu5ltEosJOJqX9qHcGBAEoDaoJTADULOJhoA7t9505UVchJPh6
taPe+vF4O3CN2yiBvLSRUXHvRjTWX0RIsADlcB9JNHumWPPqtAYvZraUImgbTbL8
O5BcoCoFmSK+No3aSuv3dR58FZMCgYEA4+oghuZ4h3EUKMY8QThp+yhS5lS/AKQu
yZzxomfx0PCO/fkWDG7cqNLXKCWZ28nlSDx1+8dUTj94jeY/mnIHMOIaH/ry3SZW
NJpIQBs4DbINrU4JLD3FWWDrPh5+JN5yKgX+14vRmFe6R38vr6AtTXT+nOaEcfE0
1OPWr5dMNcECgYB6QrARwUAdsTUBV5I/NQKkURK9ZcqaKOIVG8hPt5pF+CrCV5Xk
+wzidduE7Lp7ChGvKz+M47q7EScHBv1e61gZoZhiRmSYtxWNh740Az/DQ0XtLay5
44pBSqUM+zbGupppjOP705qDHD4OA/eo0zgAH6V70lmFViNVX8OBNvBLHwKBgCOg
5QbRnoPlzHX3T1IOxJqLmjIBi35JLDs+OpPd1fKIuIHBX44AAqStmQ7gmeW+8QXS
1crPRUGaMHlWRhkZvEALCHR5YV/q70z31VWYK7IQZIz0BwEQgvpO6VdjouqWj5g8
KbN+WvyKskcc/dJhotNZ97eFXa0GPPEO0O/QIgzBAoGAZzvNZNWF9FrEbSFlQlF6
lkufTT3THaHTYCFX9M5QkEilDPp9tDqW89u5D6HZb8ylGBaZHKvVLcG9k8jI50ys
YsIn0weTSQ1JnjDep5HkZb97neebmVPemSP+I1aaC8TqUXhZWIjK+GhLXHriQXVF
zwKpQUESvsp7DJlaOzRolak=
-----END PRIVATE KEY-----			
			`)},
			wantErr: false,
		},
		{
			name: "Invalid format",
			args: args{strings.NewReader(`
-----BEGIN PRIVATE KEY-----
123
-----END PRIVATE KEY-----			
			`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadPrivateKey(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotNil(t, got)
			}
		})
	}
}

func Test_ReadPublicKey(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid format",
			args: args{strings.NewReader(`
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1wcGldfrCtCEPE2l8Tpn
Myha7rSnzflo4Ez0H+fffhmTQhxHG/9pp7DYha6rMGjTzZ0sCN+ENbmJRL1BGeNF
XZR5qs3ez1XzywDv+6Hhao1nCoRB13Z01cvaZxE7oZcD51+kFrbwoETSdtiritcd
+oBitVAUAJcbsS/JJm/4gLiMJHfkSH2yl+6ewu/2zoEFQSU/OR2FMNLvdOXAiSwK
815Y6NzEvmpKuhUV8hLR+HWiAPekINd7/noTN4KXfBKPxFlk61y228j+dRqzJhrq
n5fp4aEBZDWClOmn3RLYkKMxhwqrJyEn54oNgaGL7OOR+rOA6McoL2VQFa2oyKey
0wIDAQAB
-----END PUBLIC KEY-----
			`)},
			wantErr: false,
		},
		{
			name: "Invalid format",
			args: args{strings.NewReader(`
-----BEGIN PUBLIC KEY-----
-----END PUBLIC KEY-----
			`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadPublicKey(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotNil(t, got)
			}
		})
	}
}
