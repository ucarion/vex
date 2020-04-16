package vex

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"math"
)

type Client interface {
	GetFlag(ctx context.Context, namespace, flagName string) (Flag, error)
}

var DefaultClient Client

type Flag struct {
	Namespace  string
	Name       string
	Expression FlagExpression
}

type FlagExpression struct {
	Type     FlagExpressionType `json:"type"`
	Constant bool               `json:"constant"`
	Percent  float64            `json:"percent"`
	ValueIn  []string           `json:"value_in"`
	Not      *FlagExpression    `json:"not"`
	AllOf    []FlagExpression   `json:"all_of"`
	AnyOf    []FlagExpression   `json:"any_of"`
	Ref      string             `json:"ref"`
}

type FlagExpressionType string

const (
	FlagExpressionTypeConstant FlagExpressionType = "constant"
	FlagExpressionTypePercent                     = "percent"
	FlagExpressionValueIn                         = "value_in"
	FlagExpressionNot                             = "not"
	FlagExpressionAllOf                           = "all_of"
	FlagExpressionAnyOf                           = "any_of"
	FlagExpressionRef                             = "ref"
)

func On(ctx context.Context, namespace, flagName, value string) (bool, error) {
	return Eval(ctx, DefaultClient, namespace, flagName, value)
}

func Eval(ctx context.Context, client Client, namespace, flagName, value string) (bool, error) {
	flag, err := client.GetFlag(ctx, namespace, flagName)
	if err != nil {
		return false, err
	}

	return eval(ctx, client, namespace, flagName, value, flag.Expression)
}

func eval(ctx context.Context, client Client, namespace, flagName, value string, expr FlagExpression) (bool, error) {
	switch expr.Type {
	case FlagExpressionTypeConstant:
		return expr.Constant, nil
	case FlagExpressionTypePercent:
		h := sha1.New()
		h.Write([]byte(value))
		n := binary.BigEndian.Uint64(h.Sum(nil))

		return n < uint64(expr.Percent*math.MaxUint64), nil
	case FlagExpressionValueIn:
		for _, v := range expr.ValueIn {
			if v == value {
				return true, nil
			}
		}

		return false, nil
	case FlagExpressionNot:
		ok, err := eval(ctx, client, namespace, flagName, value, *expr.Not)
		if err != nil {
			return false, err
		}

		return !ok, nil
	case FlagExpressionAllOf:
		for _, e := range expr.AllOf {
			ok, err := eval(ctx, client, namespace, flagName, value, e)
			if err != nil {
				return false, err
			}

			if !ok {
				return false, nil
			}
		}

		return true, nil
	case FlagExpressionAnyOf:
		for _, e := range expr.AnyOf {
			ok, err := eval(ctx, client, namespace, flagName, value, e)
			if err != nil {
				return false, err
			}

			if ok {
				return true, nil
			}
		}

		return false, nil
	case FlagExpressionRef:
		return Eval(ctx, client, expr.Ref, value, "")
	default:
		return false, errors.New("unknown expression type")
	}
}
