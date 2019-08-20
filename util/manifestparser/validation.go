package manifestparser

type validatorFunc func(ParsedManifest) error

func (m ParsedManifest) Validate() error {
	var err error

	for _, validator := range m.validators {
		err = validator(m)
		if err != nil {
			return err
		}
	}

	return nil
}
