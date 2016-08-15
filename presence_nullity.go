package meta

type Presence struct {
	Present bool
}

func (p *Presence) Optional() bool {
	return true
}

// Sets p.Present to true or false based on err and discardBlank.
// err is the output of FormValue. It could be nil or some error or ErrBlank.
// If err is ErrBlank and discardBlank, then it's like the field is not present.
// Returns the input err unless we discarded a blank, in which case it returns nil.
func (p *Presence) SetPresence(err Errorable, discardBlank bool) Errorable {
	if err == ErrBlank && discardBlank {
		p.Present = false
		return nil
	} else if err == nil {
		p.Present = true
	} else {
		p.Present = false
	}
	return err
}

type Nullity struct {
	Null bool
}

func (n *Nullity) SetNullity(err Errorable) Errorable {
	if err == ErrBlank {
		err = nil
		n.Null = true
	}
	return err
}
