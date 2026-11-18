// Package seed loads demo data so the UI looks alive on first run.
package seed

import (
	"context"
	"log"
	"time"

	"bibliotheca/internal/models"
	"bibliotheca/internal/repository"
	"bibliotheca/internal/service"
)

type bk struct {
	title, author, category, isbn string
	year, copies                  int
	desc                          string
}

var books = []bk{
	{"Atomic Habits", "James Clear", "Self-Help", "9780735211292", 2018, 4, "Tiny changes, remarkable results: a proven framework for building good habits and breaking bad ones."},
	{"The Alchemist", "Paulo Coelho", "Fiction", "9780062315007", 1988, 3, "A shepherd boy's journey to the Egyptian pyramids in search of treasure and his personal legend."},
	{"Ikigai", "Hector Garcia", "Self-Help", "9780143130727", 2016, 2, "The Japanese secret to a long and happy life."},
	{"The Psychology of Money", "Morgan Housel", "Finance", "9780857197689", 2020, 3, "Timeless lessons on wealth, greed and happiness."},
	{"Deep Work", "Cal Newport", "Productivity", "9781455586691", 2016, 2, "Rules for focused success in a distracted world."},
	{"Sapiens", "Yuval Noah Harari", "History", "9780062316110", 2011, 3, "A brief history of humankind."},
	{"1984", "George Orwell", "Classic", "9780451524935", 1949, 3, "A chilling vision of totalitarian surveillance."},
	{"The Hobbit", "J.R.R. Tolkien", "Fantasy", "9780547928227", 1937, 2, "Bilbo Baggins is swept into an epic quest to reclaim a dwarf kingdom."},
	{"Dune", "Frank Herbert", "Sci-Fi", "9780441172719", 1965, 2, "Politics, prophecy and spice on the desert planet Arrakis."},
	{"Pride and Prejudice", "Jane Austen", "Classic", "9780141439518", 1813, 2, "Elizabeth Bennet and Mr Darcy in Austen's sparkling comedy of manners."},
	{"To Kill a Mockingbird", "Harper Lee", "Classic", "9780061120084", 1960, 2, "Justice and childhood in the American South."},
	{"The Great Gatsby", "F. Scott Fitzgerald", "Classic", "9780743273565", 1925, 2, "Jazz-age glamour and the cost of chasing the American dream."},
	{"Clean Code", "Robert C. Martin", "Technology", "9780132350884", 2008, 3, "A handbook of agile software craftsmanship."},
	{"The Pragmatic Programmer", "Andrew Hunt", "Technology", "9780135957059", 2019, 2, "Your journey to mastery."},
	{"Think and Grow Rich", "Napoleon Hill", "Self-Help", "9781585424337", 1937, 2, "The classic on the mindset of achievement."},
	{"Wings of Fire", "A.P.J. Abdul Kalam", "Biography", "9788173711466", 1999, 3, "The autobiography of India's Missile Man."},
	{"The Little Prince", "Antoine de Saint-Exupery", "Fiction", "9780156012195", 1943, 2, "A poetic tale about love, loss and what matters."},
	{"Harry Potter and the Sorcerer's Stone", "J.K. Rowling", "Fantasy", "9780590353427", 1997, 4, "The boy who lived begins his magical education."},
	{"The Da Vinci Code", "Dan Brown", "Thriller", "9780307474278", 2003, 2, "A murder in the Louvre unravels a centuries-old secret."},
	{"Thinking, Fast and Slow", "Daniel Kahneman", "Psychology", "9780374533557", 2011, 2, "How two systems of thought drive our decisions."},
	{"The Midnight Library", "Matt Haig", "Fiction", "9780525559474", 2020, 3, "Between life and death there is a library of every life you could have lived."},
	{"Rich Dad Poor Dad", "Robert Kiyosaki", "Finance", "9781612680194", 1997, 3, "What the rich teach their kids about money."},
}

// Run inserts demo data when the catalogue is empty.
func Run(ctx context.Context, store repository.Store, books_ service.BookService, members service.MemberService) {
	if _, n, err := store.Books().List(ctx, models.BookFilter{PageSize: 1}); err != nil || n > 0 {
		return
	}
	log.Println("seeding demo data...")
	var ids []int64
	for _, b := range books {
		created, err := books_.Create(ctx, models.BookInput{Title: b.title, Author: b.author, Category: b.category,
			ISBN: b.isbn, PublishedYear: b.year, TotalCopies: b.copies, Description: b.desc})
		if err != nil {
			log.Printf("seed book %q: %v", b.title, err)
			continue
		}
		ids = append(ids, created.ID)
	}
	people := []models.MemberInput{
		{Name: "Anjali Menon", Email: "anjali.menon@example.com", Phone: "+91 98470 11223", Address: "Kochi, Kerala", MembershipType: "Premium"},
		{Name: "Rahul Nair", Email: "rahul.nair@example.com", Phone: "+91 98950 33441", Address: "Thrissur, Kerala", MembershipType: "Standard"},
		{Name: "Fathima Basheer", Email: "fathima.b@example.com", Phone: "+91 99610 55782", Address: "Kozhikode, Kerala", MembershipType: "Student"},
		{Name: "Arjun Krishnan", Email: "arjun.k@example.com", Phone: "+91 94470 88120", Address: "Palakkad, Kerala", MembershipType: "Standard"},
		{Name: "Meera Iyer", Email: "meera.iyer@example.com", Phone: "+91 97460 22190", Address: "Chennai, Tamil Nadu", MembershipType: "Premium"},
		{Name: "Sebastian Joseph", Email: "sebastian.j@example.com", Phone: "+91 98460 71234", Address: "Kottayam, Kerala", MembershipType: "Student"},
	}
	var mids []int64
	for _, p := range people {
		m, err := members.Create(ctx, p)
		if err != nil {
			log.Printf("seed member %q: %v", p.Name, err)
			continue
		}
		mids = append(mids, m.ID)
	}
	if len(ids) < 10 || len(mids) < 4 {
		return
	}
	now := time.Now()
	day := 24 * time.Hour
	type ln struct {
		b, m     int
		issuedAg int // days ago
		due      int // loan days
		returned int // days ago returned, -1 = still out
	}
	for _, l := range []ln{
		{0, 0, 30, 14, -1}, // overdue
		{5, 1, 20, 14, -1}, // overdue
		{1, 2, 4, 14, -1},
		{7, 3, 2, 14, -1},
		{12, 0, 6, 14, -1},
		{3, 4, 40, 14, 30},
		{8, 1, 50, 14, 40},
		{0, 3, 70, 14, 60},
	} {
		issued := now.Add(-time.Duration(l.issuedAg) * day)
		loan, err := store.Loans().Issue(ctx, models.IssueParams{BookID: ids[l.b], MemberID: mids[l.m],
			IssuedAt: issued, DueDate: issued.Add(time.Duration(l.due) * day), MaxActive: 5})
		if err != nil {
			continue
		}
		if l.returned >= 0 {
			_, _ = store.Loans().Return(ctx, loan.ID, now.Add(-time.Duration(l.returned)*day), 2)
		}
	}
}
