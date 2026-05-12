package authenticatorpb

func (x *VerifyTokenResponse) IsSingleSession() bool {
	return x.GetSingleSession()
}
