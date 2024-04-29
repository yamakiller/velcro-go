package metainfo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/yamakiller/velcro-go/utils/metainfo"
)

func calls(ctx context.Context, level int, t *testing.T, expect bool) {
	k := fmt.Sprintf("key-%d", level)
	v := fmt.Sprintf("val-%d", level)
	b := metainfo.SetBackwardValue(ctx, k, v)
	assert(t, expect == b, "expect", expect, "got", b)

	if level > 0 {
		calls(ctx, level-1, t, expect)
	}
}

func TestWithBackwardValues(t *testing.T) {
	ctx := context.Background()

	ctx = metainfo.WithBackwardValues(ctx)
	calls(ctx, 2, t, true)

	m := metainfo.GetAllBackwardValues(ctx)
	assert(t, len(m) == 3)
	assert(t, m["key-0"] == "val-0")
	assert(t, m["key-1"] == "val-1")
	assert(t, m["key-2"] == "val-2")
}

func TestWithBackwardValues2(t *testing.T) {
	ctx := context.Background()
	calls(ctx, 2, t, false)

	m := metainfo.GetAllBackwardValues(ctx)
	assert(t, len(m) == 0)
}

func TestWithBackwardValues3(t *testing.T) {
	ctx0 := context.Background()
	ctx1 := metainfo.WithBackwardValues(ctx0)
	ctx2 := metainfo.WithBackwardValues(ctx1)
	assert(t, ctx0 != ctx1)
	assert(t, ctx1 == ctx2)
}

func TestWithBackwardValues4(t *testing.T) {
	ctx0 := context.Background()
	ctx1 := metainfo.WithBackwardValues(ctx0)
	ctx2 := metainfo.WithValue(ctx1, "key", "forward")

	val, ok := metainfo.GetBackwardValue(ctx0, "key")
	assert(t, !ok)

	ok = metainfo.SetBackwardValue(ctx2, "key", "backward")
	assert(t, ok)

	val, ok = metainfo.GetValue(ctx2, "key")
	assert(t, ok)
	assert(t, val == "forward")

	val, ok = metainfo.GetBackwardValue(ctx2, "key")
	assert(t, ok)
	assert(t, val == "backward")

	val, ok = metainfo.GetBackwardValue(ctx1, "key")
	assert(t, ok)
	assert(t, val == "backward")

	ctx3 := metainfo.WithBackwardValues(ctx2)

	val, ok = metainfo.GetValue(ctx3, "key")
	assert(t, ok)
	assert(t, val == "forward")

	val, ok = metainfo.GetBackwardValue(ctx3, "key")
	assert(t, ok)
	assert(t, val == "backward")

	ok = metainfo.SetBackwardValue(ctx3, "key", "backward2")
	assert(t, ok)

	val, ok = metainfo.GetBackwardValue(ctx1, "key")
	assert(t, ok)
	assert(t, val == "backward2")
}

func TestWithBackwardValues5(t *testing.T) {
	ctx0 := context.Background()
	ctx1 := metainfo.WithBackwardValues(ctx0)
	ctx2 := metainfo.WithBackwardValuesToSend(ctx1)
	ctx3 := metainfo.WithValue(ctx2, "key", "forward")

	val, ok := metainfo.RecvBackwardValue(ctx3, "key")
	assert(t, !ok)
	assert(t, val == "")

	m := metainfo.RecvAllBackwardValues(ctx3)
	assert(t, m == nil)

	m = metainfo.AllBackwardValuesToSend(ctx3)
	assert(t, m == nil)

	ok = metainfo.SetBackwardValue(ctx0, "key", "recv")
	assert(t, !ok)

	ok = metainfo.SendBackwardValue(ctx1, "key", "send")
	assert(t, !ok)

	ok = metainfo.SetBackwardValue(ctx3, "key", "recv")
	assert(t, ok)

	ok = metainfo.SendBackwardValue(ctx3, "key", "send")
	assert(t, ok)

	ok = metainfo.SetBackwardValues(ctx3)
	assert(t, !ok)

	val, ok = metainfo.RecvBackwardValue(ctx3, "key")
	assert(t, ok && val == "recv")

	ok = metainfo.SetBackwardValues(ctx3, "key", "recv0", "key1", "recv1")
	assert(t, ok)

	ok = metainfo.SetBackwardValues(ctx3, "key", "recv2", "key1")
	assert(t, !ok)

	ok = metainfo.SendBackwardValues(ctx3)
	assert(t, !ok)

	ok = metainfo.SendBackwardValues(ctx3, "key", "send0", "key1", "send1")
	assert(t, ok)

	ok = metainfo.SendBackwardValues(ctx3, "key", "send2", "key1")
	assert(t, !ok)

	val, ok = metainfo.GetBackwardValueToSend(ctx3, "key")
	assert(t, ok)
	assert(t, val == "send0")

	m = metainfo.RecvAllBackwardValues(ctx3)
	assert(t, len(m) == 2)
	assert(t, m["key"] == "recv0" && m["key1"] == "recv1", m)

	m = metainfo.AllBackwardValuesToSend(ctx3)
	assert(t, len(m) == 2)
	assert(t, m["key"] == "send0" && m["key1"] == "send1")
}
