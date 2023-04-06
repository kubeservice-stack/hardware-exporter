package config

type SectionError string

func (e SectionError) Error() string {
	return "section not found: " + string(e)
}

type OptionError string

func (e OptionError) Error() string {
	return "option not found: " + string(e)
}
