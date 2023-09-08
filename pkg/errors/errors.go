package errors

import (
	stderrors "errors"
)

var (
	ErrUnsupportedClientVersion  = stderrors.New("unsupported client version")
	ErrNoAddresses               = stderrors.New("no addresses")
	ErrInvalidAddress            = stderrors.New("invalid address")
	ErrUnsetSource               = stderrors.New("unset source")
	ErrInvalidIndex              = stderrors.New("invalid index")
	ErrNoSpaceName               = stderrors.New("no space name")
	ErrNoGraphName               = stderrors.New("no graph name")
	ErrNoNodeName                = stderrors.New("no node name")
	ErrNoNodeID                  = stderrors.New("no node id")
	ErrNoEdgeSrc                 = stderrors.New("no edge src")
	ErrNoEdgeDst                 = stderrors.New("no edge dst")
	ErrNoEdgeName                = stderrors.New("no edge name")
	ErrNoNodeIDName              = stderrors.New("no node id name")
	ErrNoPropName                = stderrors.New("no prop name")
	ErrNoProps                   = stderrors.New("no props")
	ErrUnsupportedValueType      = stderrors.New("unsupported value type")
	ErrNoRecord                  = stderrors.New("no record")
	ErrNoIndicesOrConcatItems    = stderrors.New("no indices or concat items")
	ErrUnsupportedConcatItemType = stderrors.New("unsupported concat item type")
	ErrUnsupportedFunction       = stderrors.New("unsupported function")
	ErrFilterSyntax              = stderrors.New("filter syntax")
	ErrUnsupportedMode           = stderrors.New("unsupported mode")
)
