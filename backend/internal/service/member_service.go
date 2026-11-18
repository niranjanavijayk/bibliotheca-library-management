package service

import (
	"context"
	"strings"

	"bibliotheca/internal/apperror"
	"bibliotheca/internal/models"
	"bibliotheca/internal/repository"
)

// MemberService manages library members.
type MemberService interface {
	Create(ctx context.Context, in models.MemberInput) (*models.Member, error)
	Update(ctx context.Context, id int64, in models.MemberInput) (*models.Member, error)
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*models.Member, error)
	List(ctx context.Context, f models.MemberFilter) ([]models.Member, int, error)
}

type memberService struct {
	members repository.MemberRepository
	clock   Clock
}

func NewMemberService(members repository.MemberRepository, clock Clock) MemberService {
	return &memberService{members: members, clock: clock}
}

func (s *memberService) validate(ctx context.Context, in *models.MemberInput, excludeID int64) error {
	fe := fieldErrors{}
	in.Name = fe.required("name", "Name", in.Name, 150)
	in.Email = strings.ToLower(fe.required("email", "Email", in.Email, 190))
	if in.Email != "" && !emailRe.MatchString(in.Email) {
		fe.add("email", "Enter a valid email address")
	}
	in.Phone = fe.required("phone", "Phone", in.Phone, 25)
	if in.Phone != "" && !phoneRe.MatchString(in.Phone) {
		fe.add("phone", "Enter a valid phone number (7-15 digits)")
	}
	in.Address = strings.TrimSpace(in.Address)
	if in.MembershipType == "" {
		in.MembershipType = "Standard"
	}
	valid := false
	for _, t := range models.MembershipTypes {
		if t == in.MembershipType {
			valid = true
		}
	}
	if !valid {
		fe.add("membershipType", "Choose Standard, Premium or Student")
	}
	if in.Status == "" {
		in.Status = models.MemberActive
	}
	if in.Status != models.MemberActive && in.Status != models.MemberSuspended {
		fe.add("status", "Status must be active or suspended")
	}
	if err := fe.err(); err != nil {
		return err
	}
	exists, err := s.members.EmailExists(ctx, in.Email, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return apperror.Validation(map[string]string{"email": "A member with this email already exists"})
	}
	return nil
}

func (s *memberService) Create(ctx context.Context, in models.MemberInput) (*models.Member, error) {
	if err := s.validate(ctx, &in, 0); err != nil {
		return nil, err
	}
	now := s.clock()
	m := &models.Member{Name: in.Name, Email: in.Email, Phone: in.Phone, Address: in.Address,
		MembershipType: in.MembershipType, Status: in.Status, JoinedAt: now, ExpiresAt: now.AddDate(1, 0, 0)}
	if err := s.members.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *memberService) Update(ctx context.Context, id int64, in models.MemberInput) (*models.Member, error) {
	cur, err := s.members.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.validate(ctx, &in, id); err != nil {
		return nil, err
	}
	m := &models.Member{ID: id, Name: in.Name, Email: in.Email, Phone: in.Phone, Address: in.Address,
		MembershipType: in.MembershipType, Status: in.Status, ExpiresAt: cur.ExpiresAt}
	// Re-activating a lapsed member renews their membership for a year.
	if m.Status == models.MemberActive && m.ExpiresAt.Before(s.clock()) {
		m.ExpiresAt = s.clock().AddDate(1, 0, 0)
	}
	if err := s.members.Update(ctx, m); err != nil {
		return nil, err
	}
	return s.members.GetByID(ctx, id)
}

func (s *memberService) Delete(ctx context.Context, id int64) error {
	m, err := s.members.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if m.ActiveLoans > 0 {
		return apperror.Conflict("This member still has books on loan. Process the returns first.")
	}
	return s.members.Delete(ctx, id)
}

func (s *memberService) Get(ctx context.Context, id int64) (*models.Member, error) {
	return s.members.GetByID(ctx, id)
}

func (s *memberService) List(ctx context.Context, f models.MemberFilter) ([]models.Member, int, error) {
	f.Query = strings.TrimSpace(f.Query)
	return s.members.List(ctx, f)
}
