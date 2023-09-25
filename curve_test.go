package starknetgo

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/NethermindEth/starknet.go/types"
)

func BenchmarkPedersenHash(b *testing.B) {
	suite := [][]*big.Int{
		{types.HexToBN("0x12773"), types.HexToBN("0x872362")},
		{types.HexToBN("0x1277312773"), types.HexToBN("0x872362872362")},
		{types.HexToBN("0x1277312773"), types.HexToBN("0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826")},
		{types.HexToBN("0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB"), types.HexToBN("0x872362872362")},
		{types.HexToBN("0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826"), types.HexToBN("0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB")},
		{types.HexToBN("0x7f15c38ea577a26f4f553282fcfe4f1feeb8ecfaad8f221ae41abf8224cbddd"), types.HexToBN("0x13d41f388b8ea4db56c5aa6562f13359fab192b3db57651af916790f9debee9")},
		{types.HexToBN("0x7f15c38ea577a26f4f553282fcfe4f1feeb8ecfaad8f221ae41abf8224cbddd"), types.HexToBN("0x7f15c38ea577a26f4f553282fcfe4f1feeb8ecfaad8f221ae41abf8224cbdde")},
	}

	for _, test := range suite {
		b.Run(fmt.Sprintf("input_size_%d_%d", test[0].BitLen(), test[1].BitLen()), func(b *testing.B) {
			Curve.PedersenHash(test)
		})
	}
}

func BenchmarkCurveSign(b *testing.B) {
	type data struct {
		MessageHash *big.Int
		PrivateKey  *big.Int
		Seed        *big.Int
	}

	dataSet := []data{}
	MessageHash := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(250), nil)
	PrivateKey := big.NewInt(0).Add(MessageHash, big.NewInt(1))
	Seed := big.NewInt(0)
	for i := int64(0); i < 20; i++ {
		dataSet = append(dataSet, data{
			MessageHash: big.NewInt(0).Add(MessageHash, big.NewInt(i)),
			PrivateKey:  big.NewInt(0).Add(PrivateKey, big.NewInt(i)),
			Seed:        big.NewInt(0).Add(Seed, big.NewInt(i)),
		})

		for _, test := range dataSet {
			Curve.Sign(test.MessageHash, test.PrivateKey, test.Seed)
		}
	}
}

func TestGeneral_PrivateToPoint(t *testing.T) {
	x, _, err := Curve.PrivateToPoint(big.NewInt(2))
	if err != nil {
		t.Errorf("PrivateToPoint err %v", err)
	}
	expectedX, _ := new(big.Int).SetString("3324833730090626974525872402899302150520188025637965566623476530814354734325", 10)
	if x.Cmp(expectedX) != 0 {
		t.Errorf("Actual public key %v different from expected %v", x, expectedX)
	}
}

func TestGeneral_PedersenHash(t *testing.T) {
	testPedersen := []struct {
		elements []*big.Int
		expected *big.Int
	}{
		{
			elements: []*big.Int{types.HexToBN("0x12773"), types.HexToBN("0x872362")},
			expected: types.HexToBN("0x5ed2703dfdb505c587700ce2ebfcab5b3515cd7e6114817e6026ec9d4b364ca"),
		},
		{
			elements: []*big.Int{types.HexToBN("0x13d41f388b8ea4db56c5aa6562f13359fab192b3db57651af916790f9debee9"), types.HexToBN("0x537461726b4e6574204d61696c")},
			expected: types.HexToBN("0x180c0a3d13c1adfaa5cbc251f4fc93cc0e26cec30ca4c247305a7ce50ac807c"),
		},
		{
			elements: []*big.Int{big.NewInt(100), big.NewInt(1000)},
			expected: types.HexToBN("0x45a62091df6da02dce4250cb67597444d1f465319908486b836f48d0f8bf6e7"),
		},
	}

	for _, tt := range testPedersen {
		hash, err := Curve.PedersenHash(tt.elements)
		if err != nil {
			t.Errorf("Hashing err: %v\n", err)
		}
		if hash.Cmp(tt.expected) != 0 {
			t.Errorf("incorrect hash: got %v expected %v\n", hash, tt.expected)
		}
	}
}

func TestGeneral_DivMod(t *testing.T) {
	testDivmod := []struct {
		x        *big.Int
		y        *big.Int
		expected *big.Int
	}{
		{
			x:        types.StrToBig("311379432064974854430469844112069886938521247361583891764940938105250923060"),
			y:        types.StrToBig("621253665351494585790174448601059271924288186997865022894315848222045687999"),
			expected: types.StrToBig("2577265149861519081806762825827825639379641276854712526969977081060187505740"),
		},
		{
			x:        big.NewInt(1),
			y:        big.NewInt(2),
			expected: types.HexToBN("0x0400000000000008800000000000000000000000000000000000000000000001"),
		},
	}

	for _, tt := range testDivmod {
		divR := DivMod(tt.x, tt.y, Curve.P)

		if divR.Cmp(tt.expected) != 0 {
			t.Errorf("DivMod Res %v does not == expected %v\n", divR, tt.expected)
		}
	}
}

func TestGeneral_Add(t *testing.T) {
	testAdd := []struct {
		x         *big.Int
		y         *big.Int
		expectedX *big.Int
		expectedY *big.Int
	}{
		{
			x:         types.StrToBig("1468732614996758835380505372879805860898778283940581072611506469031548393285"),
			y:         types.StrToBig("1402551897475685522592936265087340527872184619899218186422141407423956771926"),
			expectedX: types.StrToBig("2573054162739002771275146649287762003525422629677678278801887452213127777391"),
			expectedY: types.StrToBig("3086444303034188041185211625370405120551769541291810669307042006593736192813"),
		},
		{
			x:         big.NewInt(1),
			y:         big.NewInt(2),
			expectedX: types.StrToBig("225199957243206662471193729647752088571005624230831233470296838210993906468"),
			expectedY: types.StrToBig("190092378222341939862849656213289777723812734888226565973306202593691957981"),
		},
	}

	for _, tt := range testAdd {
		resX, resY := Curve.Add(Curve.Gx, Curve.Gy, tt.x, tt.y)
		if resX.Cmp(tt.expectedX) != 0 {
			t.Errorf("ResX %v does not == expected %v\n", resX, tt.expectedX)

		}
		if resY.Cmp(tt.expectedY) != 0 {
			t.Errorf("ResY %v does not == expected %v\n", resY, tt.expectedY)
		}
	}
}

func TestGeneral_MultAir(t *testing.T) {
	testMult := []struct {
		r         *big.Int
		x         *big.Int
		y         *big.Int
		expectedX *big.Int
		expectedY *big.Int
	}{
		{
			r:         types.StrToBig("2458502865976494910213617956670505342647705497324144349552978333078363662855"),
			x:         types.StrToBig("1468732614996758835380505372879805860898778283940581072611506469031548393285"),
			y:         types.StrToBig("1402551897475685522592936265087340527872184619899218186422141407423956771926"),
			expectedX: types.StrToBig("182543067952221301675635959482860590467161609552169396182763685292434699999"),
			expectedY: types.StrToBig("3154881600662997558972388646773898448430820936643060392452233533274798056266"),
		},
	}

	for _, tt := range testMult {
		x, y, err := Curve.MimicEcMultAir(tt.r, tt.x, tt.y, Curve.Gx, Curve.Gy)
		if err != nil {
			t.Errorf("MultAirERR %v\n", err)
		}

		if x.Cmp(tt.expectedX) != 0 {
			t.Errorf("ResX %v does not == expected %v\n", x, tt.expectedX)

		}
		if y.Cmp(tt.expectedY) != 0 {
			t.Errorf("ResY %v does not == expected %v\n", y, tt.expectedY)
		}
	}
}
