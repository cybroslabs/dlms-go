package dlmsal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type AssociationResult byte

const (
	AssociationResultAccepted          AssociationResult = 0
	AssociationResultPermanentRejected AssociationResult = 1
	AssociationResultTransientRejected AssociationResult = 2
)

type SourceDiagnostic byte

const (
	SourceDiagnosticNone                                       SourceDiagnostic = 0
	SourceDiagnosticNoReasonGiven                              SourceDiagnostic = 1
	SourceDiagnosticApplicationContextNameNotSupported         SourceDiagnostic = 2
	SourceDiagnosticCallingAPTitleNotRecognized                SourceDiagnostic = 3
	SourceDiagnosticCallingAPInvocationIdentifierNotRecognized SourceDiagnostic = 4
	SourceDiagnosticCallingAEQualifierNotRecognized            SourceDiagnostic = 5
	SourceDiagnosticCallingAEInvocationIdentifierNotRecognized SourceDiagnostic = 6
	SourceDiagnosticCalledAPTitleNotRecognized                 SourceDiagnostic = 7
	SourceDiagnosticCalledAPInvocationIdentifierNotRecognized  SourceDiagnostic = 8
	SourceDiagnosticCalledAEQualifierNotRecognized             SourceDiagnostic = 9
	SourceDiagnosticCalledAEInvocationIdentifierNotRecognized  SourceDiagnostic = 10
	SourceDiagnosticAuthenticationMechanismNameNotRecognized   SourceDiagnostic = 11
	SourceDiagnosticAuthenticationMechanismNameRequired        SourceDiagnostic = 12
	SourceDiagnosticAuthenticationFailure                      SourceDiagnostic = 13
	SourceDiagnosticAuthenticationRequired                     SourceDiagnostic = 14
)

type initiateResponse struct {
	NegotiatedQualityOfService *byte
	NegotiatedConformance      uint32
	ServerMaxReceivePduSize    uint16
	VAAddress                  int16
}

type confirmedServiceErrorTag byte

const (
	TagErrInitiateError confirmedServiceErrorTag = 1
	TagErrRead          confirmedServiceErrorTag = 5
	TagErrWrite         confirmedServiceErrorTag = 6
)

type serviceErrorTag byte

const (
	TagErrApplicationReference serviceErrorTag = 0
	TagErrHardwareResource     serviceErrorTag = 1
	TagErrVdeStateError        serviceErrorTag = 2
	TagErrService              serviceErrorTag = 3
	TagErrDefinition           serviceErrorTag = 4
	TagErrAccess               serviceErrorTag = 5
	TagErrInitiate             serviceErrorTag = 6
	TagErrLoadDataSet          serviceErrorTag = 7
	TagErrTask                 serviceErrorTag = 9
	TagErrOtherError           serviceErrorTag = 10
)

type confirmedServiceError struct {
	ConfirmedServiceError confirmedServiceErrorTag
	ServiceError          serviceErrorTag
	Value                 byte
}

type ApplicationContext byte

// Application context definitions
const (
	ApplicationContextLNNoCiphering ApplicationContext = 1
	ApplicationContextSNNoCiphering ApplicationContext = 2
	ApplicationContextLNCiphering   ApplicationContext = 3
	ApplicationContextSNCiphering   ApplicationContext = 4
)

const (
	PduTypeProtocolVersion            = 0
	PduTypeApplicationContextName     = 1
	PduTypeCalledAPTitle              = 2
	PduTypeCalledAEQualifier          = 3
	PduTypeCalledAPInvocationID       = 4
	PduTypeCalledAEInvocationID       = 5
	PduTypeCallingAPTitle             = 6
	PduTypeCallingAEQualifier         = 7
	PduTypeCallingAPInvocationID      = 8
	PduTypeCallingAEInvocationID      = 9
	PduTypeSenderAcseRequirements     = 10
	PduTypeMechanismName              = 11
	PduTypeCallingAuthenticationValue = 12
	PduTypeImplementationInformation  = 29
	PduTypeUserInformation            = 30
)

const (
	BERTypeContext     = 0x80
	BERTypeApplication = 0x40
	BERTypeConstructed = 0x20
)

// Conformance block
const (
	ConformanceBlockReservedZero                = 0b100000000000000000000000
	ConformanceBlockGeneralProtection           = 0b010000000000000000000000
	ConformanceBlockGeneralBlockTransfer        = 0b001000000000000000000000
	ConformanceBlockRead                        = 0b000100000000000000000000
	ConformanceBlockWrite                       = 0b000010000000000000000000
	ConformanceBlockUnconfirmedWrite            = 0b000001000000000000000000
	ConformanceBlockReservedSix                 = 0b000000100000000000000000
	ConformanceBlockReservedSeven               = 0b000000010000000000000000
	ConformanceBlockAttribute0SupportedWithSet  = 0b000000001000000000000000
	ConformanceBlockPriorityMgmtSupported       = 0b000000000100000000000000
	ConformanceBlockAttribute0SupportedWithGet  = 0b000000000010000000000000
	ConformanceBlockBlockTransferWithGetOrRead  = 0b000000000001000000000000
	ConformanceBlockBlockTransferWithSetOrWrite = 0b000000000000100000000000
	ConformanceBlockBlockTransferWithAction     = 0b000000000000010000000000
	ConformanceBlockMultipleReferences          = 0b000000000000001000000000
	ConformanceBlockInformationReport           = 0b000000000000000100000000
	ConformanceBlockDataNotification            = 0b000000000000000010000000
	ConformanceBlockAccess                      = 0b000000000000000001000000
	ConformanceBlockParametrizedAccess          = 0b000000000000000000100000
	ConformanceBlockGet                         = 0b000000000000000000010000
	ConformanceBlockSet                         = 0b000000000000000000001000
	ConformanceBlockSelectiveAccess             = 0b000000000000000000000100
	ConformanceBlockEventNotification           = 0b000000000000000000000010
	ConformanceBlockAction                      = 0b000000000000000000000001
)

type aaretag struct {
	tag  byte
	data []byte
}

type aareresponse struct {
	ApplicationContextName ApplicationContext
	AssociationResult      AssociationResult
	SourceDiagnostic       SourceDiagnostic
	SystemTitle            []byte
	initiateResponse       *initiateResponse
	confirmedServiceError  *confirmedServiceError
}

func putappctxname(dst *bytes.Buffer, settings *DlmsSettings) {
	// not so exactly correct things, but for speed sake
	dst.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeApplicationContextName)
	dst.Write([]byte{0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01})
	switch settings.ApplicationContext {
	case ApplicationContextLNNoCiphering:
		dst.WriteByte(byte(ApplicationContextLNNoCiphering))
	case ApplicationContextSNNoCiphering:
		dst.WriteByte(byte(ApplicationContextSNNoCiphering))
	default:
		panic("unsupported")
	}
}

func putmechname(dst *bytes.Buffer, settings *DlmsSettings) {
	if settings.Authentication == AuthenticationNone {
		return
	}
	dst.WriteByte(BERTypeContext | PduTypeMechanismName)
	dst.Write([]byte{0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02})
	dst.WriteByte(byte(settings.Authentication))
}

func putsecvalues(dst *bytes.Buffer, settings *DlmsSettings) {
	if settings.Authentication == AuthenticationNone {
		return
	} // todo sec values
	encodetag2(dst, BERTypeContext|BERTypeConstructed|PduTypeCallingAuthenticationValue, 0x80, settings.Password)
}

func createxdlms(dst *bytes.Buffer, settings *DlmsSettings) {
	xdlms := make([]byte, 14)
	xdlms[0] = 0x01
	xdlms[1] = 0x00
	xdlms[2] = 0x00
	xdlms[3] = 0x00
	xdlms[4] = 0x06
	xdlms[5] = 0x5f
	xdlms[6] = 0x1f
	xdlms[7] = 0x04
	xdlms[8] = 0x00 // overwritten, but who cares
	// put conform
	binary.BigEndian.PutUint32(xdlms[8:], uint32(settings.ConformanceBlock))
	xdlms[12] = byte(settings.MaxPduRecvSize >> 8) // no limit in maximum received apdu length
	xdlms[13] = byte(settings.MaxPduRecvSize)

	encodetag2(dst, BERTypeContext|BERTypeConstructed|PduTypeUserInformation, 0x04, xdlms)
}

func encodeaarq(settings *DlmsSettings) (out []byte, err error) {
	var buf bytes.Buffer
	var content bytes.Buffer

	putappctxname(&content, settings)
	encodetag(&content, BERTypeContext|PduTypeSenderAcseRequirements, []byte{0x07, 0x80})
	putmechname(&content, settings)
	putsecvalues(&content, settings)
	createxdlms(&content, settings)

	encodetag(&buf, byte(TagAARQ), content.Bytes())
	out = buf.Bytes()
	return
}

func decodeaare(src []byte, tmp []byte) ([]aaretag, error) {
	ret := make([]aaretag, 0, 20)
	for len(src) > 0 {
		tag, l, data, err := decodetag(src, tmp)
		if err != nil {
			return nil, err
		}
		ret = append(ret, aaretag{tag: tag, data: data})
		src = src[l:]
	}
	return ret, nil
}

func parseApplicationContextName(tag *aaretag) (out ApplicationContext, err error) {
	if len(tag.data) != 9 {
		err = fmt.Errorf("invalid A1 tag length")
		return
	}
	rsp := []byte{0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01}
	if !bytes.Equal(tag.data[:8], rsp) {
		err = fmt.Errorf("invalid A1 tag content")
		return
	}
	out = ApplicationContext(tag.data[8])
	return
}

func parseAssociationResult(tag *aaretag) (out AssociationResult, err error) {
	if len(tag.data) != 3 {
		err = fmt.Errorf("invalid A2 tag length")
		return
	}
	rsp := []byte{0x02, 0x01}
	if !bytes.Equal(tag.data[:2], rsp) {
		err = fmt.Errorf("invalid A2 tag content")
		return
	}
	out = AssociationResult(tag.data[2])
	return
}

func parseAssociateSourceDiagnostic(tag *aaretag) (out SourceDiagnostic, err error) {
	if len(tag.data) != 5 {
		err = fmt.Errorf("invalid A3 tag length")
		return
	}
	rsp := []byte{0x03, 0x02, 0x01}
	if !bytes.Equal(tag.data[1:4], rsp) {
		err = fmt.Errorf("invalid A3 tag content")
		return
	}
	out = SourceDiagnostic(tag.data[4])
	return
}

func parseAPTitle(tag *aaretag, tmp []byte) (out []byte, err error) {
	if len(tag.data) < 2 {
		return nil, fmt.Errorf("invalid A4 tag length")
	}
	t, _, d, err := decodetag(tag.data, tmp)
	if err != nil {
		return nil, err
	}
	if t != 0x04 {
		return nil, fmt.Errorf("invalid A4 tag content")
	}
	out = make([]byte, len(d))
	copy(out, d)
	return
}

func parseUserInformation(tag *aaretag, tmp []byte) (ir *initiateResponse, cse *confirmedServiceError, err error) {
	if len(tag.data) < 6 {
		err = fmt.Errorf("invalid BE tag length")
		return
	}
	t, _, d, err := decodetag(tag.data, tmp)
	if err != nil {
		return nil, nil, err
	}
	if t != 0x04 {
		return nil, nil, fmt.Errorf("invalid BE tag content")
	}

	if d[0] == byte(TagInitiateResponse) {
		ir, err := decodeInitiateResponse(d)
		return &ir, nil, err
	}

	if d[0] == byte(TagConfirmedServiceError) {
		cse, err := decodeConfirmedServiceError(d)
		return nil, &cse, err
	}

	err = errors.New("unexpected user information tag")
	return
}

func decodeInitiateResponse(src []byte) (out initiateResponse, err error) {
	if len(src) < 14 {
		if len(src) == 13 { // some units can return this shit, underlying array should be big enough to accomodate additional byte
			src = src[:14] // this hack wont work if 0xbe tag is not the last one, ok, usually is the last one
		} else {
			err = fmt.Errorf("invalid initial response length")
			return
		}
	}

	if src[0] != byte(TagInitiateResponse) {
		err = fmt.Errorf("invalid initial response tag")
		return
	}

	if src[1] == 0x01 {
		negotiatedQualityOfService := src[2]
		out.NegotiatedQualityOfService = &negotiatedQualityOfService
		src = src[3:]
	} else {
		src = src[2:]
	}

	if src[0] != DlmsVersion {
		err = fmt.Errorf("wrong dlms version")
		return
	}

	if !bytes.Equal(src[1:5], []byte{0x5F, 0x1F, 0x04, 0x00}) {
		err = fmt.Errorf("invalid initial response content")
		return
	}

	out.NegotiatedConformance = binary.BigEndian.Uint32(src[4:8])
	out.ServerMaxReceivePduSize = binary.BigEndian.Uint16(src[8:10])
	out.VAAddress = int16(binary.BigEndian.Uint16(src[10:12]))
	return
}

func decodeConfirmedServiceError(src []byte) (out confirmedServiceError, err error) {
	if len(src) < 4 {
		err = fmt.Errorf("invalid service error length")
		return
	}

	if src[0] != byte(TagConfirmedServiceError) {
		err = fmt.Errorf("invalid service error tag")
		return
	}

	out.ConfirmedServiceError = confirmedServiceErrorTag(src[1])
	out.ServiceError = serviceErrorTag(src[2])
	out.Value = src[3]
	return
}
