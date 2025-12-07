package main

import (
	"github.com/bobnomicon/lenslocked/email"
	"github.com/bobnomicon/lenslocked/templates"
	"github.com/bobnomicon/lenslocked/views"
)

func parseTemplates(s *server) {
	// Parse static templates
	s.Controllers.Static.Templates.Home = views.Must(views.ParseFS(templates.FS,
		"home.gohtml", "layout.gohtml",
	))
	s.Controllers.Static.Templates.Contact = views.Must(views.ParseFS(templates.FS,
		"contact.gohtml", "layout.gohtml",
	))
	s.Controllers.Static.Templates.FAQ = views.Must(views.ParseFS(templates.FS,
		"faq.gohtml", "layout.gohtml",
	))

	// Parse users templates
	s.Controllers.Users.Templates.SignUp = views.Must(views.ParseFS(templates.FS,
		"signup.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.SignIn = views.Must(views.ParseFS(templates.FS,
		"signin.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.ForgotPassword = views.Must(views.ParseFS(templates.FS,
		"forgot-password.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.CheckEmail = views.Must(views.ParseFS(templates.FS,
		"check-email.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.ResetPassword = views.Must(views.ParseFS(templates.FS,
		"reset-password.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.ForgotPasswordEmail = email.Must(email.ParseFS(templates.FS,
		"emails/forgot-password.gohtml",
	))

	// Parse galleries templates
	s.Controllers.Galleries.Templates.New = views.Must(views.ParseFS(templates.FS,
		"galleries/new.gohtml", "layout.gohtml",
	))
	s.Controllers.Galleries.Templates.Edit = views.Must(views.ParseFS(templates.FS,
		"galleries/edit.gohtml", "layout.gohtml",
	))
	s.Controllers.Galleries.Templates.Index = views.Must(views.ParseFS(templates.FS,
		"galleries/index.gohtml", "layout.gohtml",
	))
	s.Controllers.Galleries.Templates.Show = views.Must(views.ParseFS(templates.FS,
		"galleries/show.gohtml", "layout.gohtml",
	))
}
