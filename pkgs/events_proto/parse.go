package events_proto

import "strconv"

// NewAttributeFromString builds an Attribute from key and a string value.
// The value is parsed at runtime: we try bool, then int64, then float64;
// if any parse fails (e.g. overflow or invalid), we keep the value as string.
// So "true"/"false" become bool, "42" becomes int64, "3.14" becomes double,
// and anything else (including too-large numbers) stays as string_value.
func NewAttributeFromString(key, value string) *Attribute {
	return &Attribute{
		Key:   key,
		Value: parseAttributeValue(value),
	}
}

// parseAttributeValue tries to interpret s as bool, int64, or float64, in that order.
// On failure (including overflow), it returns a string_value so the value is never lost.
func parseAttributeValue(s string) isAttribute_Value {
	if b, ok := parseBool(s); ok {
		return &Attribute_BoolValue{BoolValue: b}
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &Attribute_Int64Value{Int64Value: i}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return &Attribute_DoubleValue{DoubleValue: f}
	}
	return &Attribute_StringValue{StringValue: s}
}

func parseBool(s string) (bool, bool) {
	switch s {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false // return false and false if the value is not a boolean
	}
}
