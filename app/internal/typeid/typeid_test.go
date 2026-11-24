package typeid

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func uuidOf(t *testing.T, text string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(text); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestSpecVectors(t *testing.T) {
	vectors := map[string]string{
		"00000000-0000-0000-0000-000000000000": "00000000000000000000000000",
		"00000000-0000-0000-0000-000000000001": "00000000000000000000000001",
		"00000000-0000-0000-0000-000000000020": "00000000000000000000000010",
		"0110c853-1d09-52d8-d73e-1194e95b5f19": "0123456789abcdefghjkmnpqrs",
		"01890a5d-ac96-774b-bcce-b302099a8057": "01h455vb4pex5vsknk084sn02q",
		"ffffffff-ffff-ffff-ffff-ffffffffffff": "7zzzzzzzzzzzzzzzzzzzzzzzzz",
	}
	for uuid, suffix := range vectors {
		id := uuidOf(t, uuid)
		encoded := Encode("msg", id.Bytes)
		if encoded != "msg_"+suffix {
			t.Errorf("Encode(%s) = %s, want msg_%s", uuid, encoded, suffix)
		}
		decoded, err := Decode("msg", encoded)
		if err != nil || decoded != id.Bytes {
			t.Errorf("Decode(%s) = %x, %v", encoded, decoded, err)
		}
	}
}

func TestDecodeRejects(t *testing.T) {
	for _, value := range []string{
		"msg_8zzzzzzzzzzzzzzzzzzzzzzzzz",
		"ses_01h455vb4pex5vsknk084sn02q",
		"msg_01h455vb4pex5vsknk084sn02",
		"msg_01h455vb4pex5vsknk084sn0iq",
		"01h455vb4pex5vsknk084sn02q",
	} {
		if _, err := Decode("msg", value); err == nil {
			t.Errorf("Decode(%q) accepted", value)
		}
	}
}

func TestParseMessageIDAcceptsLegacyUUID(t *testing.T) {
	legacy := "01890a5d-ac96-774b-bcce-b302099a8057"
	id, err := ParseMessageID(legacy)
	if err != nil || id != uuidOf(t, legacy) {
		t.Fatalf("ParseMessageID(%s) = %v, %v", legacy, id, err)
	}
	if got := NormalizeMessageID(legacy); got != "msg_01h455vb4pex5vsknk084sn02q" {
		t.Fatalf("NormalizeMessageID = %s", got)
	}
}
