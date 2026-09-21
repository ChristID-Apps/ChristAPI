package routes

import (
	"christ-api/internal/activities"
	"christ-api/internal/activitysubmissions"
	"christ-api/internal/attendance"
	"christ-api/internal/auth"
	"christ-api/internal/bible"
	"christ-api/internal/contacts"
	"christ-api/internal/middleware"
	"christ-api/internal/news"
	"christ-api/internal/points"
	"christ-api/internal/rewards"
	"christ-api/internal/role"
	"christ-api/internal/sites"
	"christ-api/internal/streaks"
	"christ-api/pkg/database"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	// Initialize handlers with database
	authHandler := auth.NewHandler(&auth.AuthRepository{DB: database.DB})
	contactsHandler := contacts.NewHandler(&contacts.ContactRepository{DB: database.DB})
	activityHandler := activities.NewHandler(&activities.Repository{DB: database.DB})
	submissionHandler := activitysubmissions.NewHandler(&activitysubmissions.Repository{DB: database.DB})
	roleRepository := &role.RoleRepository{DB: database.DB}
	roleHandler := role.NewHandler(roleRepository)
	pointsHandler := points.NewHandler(&points.Repository{DB: database.DB})
	attendanceHandler := attendance.NewHandler(&attendance.Repository{DB: database.DB})
	streakHandler := streaks.NewHandler(&streaks.Repository{DB: database.DB})
	newsHandler := news.NewHandler(&news.NewsRepository{DB: database.DB})
	rewardsHandler := rewards.NewHandler(&rewards.Repository{DB: database.DB})
	sitesHandler := sites.NewHandler(&sites.SiteRepository{DB: database.DB})
	bibleHandler := bible.NewHandler(&bible.BibleRepository{DB: database.DB})

	api := app.Group("/api")

	// Bible read routes
	api.Get("/bible/versions", bibleHandler.ListVersions)
	api.Get("/bible/books", bibleHandler.ListBooks)
	api.Get("/bible/:version/search", bibleHandler.Search)
	api.Get("/bible/:version/:book/:chapter/:verse", bibleHandler.GetVerse)
	api.Get("/bible/:version/:book/:chapter", bibleHandler.GetChapter)

	// public auth routes
	api.Post("/login", authHandler.Login)
	api.Post("/register", authHandler.Register)
	api.Post("/verify-otp", authHandler.VerifyOTP)
	api.Post("/auth/google", authHandler.LoginGoogle)
	api.Post("/auth/google/username", authHandler.SubmitGoogleUsername)

	// protected routes
	protected := api.Group("/", middleware.AuthMiddleware)

	protected.Get("/profile", contactsHandler.MyProfile)
	protected.Patch("/profile", contactsHandler.UpdateMyProfile)
	protected.Post("/profile/photo", contactsHandler.UploadProfilePhoto)

	protected.Post("/logout", authHandler.Logout)

	// admin approvals (admin only)
	adminService := &role.RoleService{Repo: roleRepository}
	adminRoutes := protected.Group("/admin", middleware.AdminOnly(adminService))
	adminRoutes.Get("/approvals", authHandler.GetPendingApprovals)
	adminRoutes.Post("/approvals/:id/approve", authHandler.ApproveUser)
	adminRoutes.Post("/approvals/:id/reject", authHandler.RejectUser)

	// roles (admin only)
	adminRoutes.Get("/roles", roleHandler.List)
	adminRoutes.Post("/roles", roleHandler.Create)
	adminRoutes.Patch("/roles/:id", roleHandler.Update)

	// activities (admin only)
	adminRoutes.Post("/activities", activityHandler.Create)
	adminRoutes.Patch("/activities/:uuid", activityHandler.Update)
	adminRoutes.Delete("/activities/:uuid", activityHandler.Delete)
	adminRoutes.Post("/activities/:uuid/image", activityHandler.UploadImage)
	adminRoutes.Patch("/activities/:uuid/bible-config", activityHandler.ConfigureBible)
	adminRoutes.Get("/activity-submissions", submissionHandler.AdminList)
	adminRoutes.Post("/activity-submissions/:uuid/approve", submissionHandler.Approve)
	adminRoutes.Post("/activity-submissions/:uuid/reject", submissionHandler.Reject)

	// sites (admin only)
	adminRoutes.Post("/sites", sitesHandler.Create)
	adminRoutes.Patch("/sites/:uuid", sitesHandler.Update)

	// contacts (admin only)
	adminRoutes.Post("/contacts", contactsHandler.Create)
	adminRoutes.Patch("/contacts/:id", contactsHandler.Update)
	adminRoutes.Delete("/contacts/:id", contactsHandler.Delete)
	adminRoutes.Get("/contacts", contactsHandler.List)
	adminRoutes.Get("/contacts/:id", contactsHandler.List)

	// points (admin only)
	adminRoutes.Get("/points", pointsHandler.Get)
	adminRoutes.Post("/points/earn", pointsHandler.Earn)
	adminRoutes.Get("/rewards", rewardsHandler.AdminList)
	adminRoutes.Post("/rewards", rewardsHandler.Create)
	adminRoutes.Patch("/rewards/:uuid", rewardsHandler.Update)
	adminRoutes.Post("/rewards/:uuid/image", rewardsHandler.UploadImage)
	adminRoutes.Get("/reward-redemptions", rewardsHandler.AdminRedemptions)
	adminRoutes.Post("/reward-redemptions/:uuid/approve", rewardsHandler.Approve)
	adminRoutes.Post("/reward-redemptions/:uuid/reject", rewardsHandler.Reject)
	adminRoutes.Post("/reward-redemptions/:uuid/complete", rewardsHandler.Complete)

	// streaks (admin only)
	adminRoutes.Post("/news", newsHandler.Create)
	adminRoutes.Patch("/news/:uuid", newsHandler.Update)
	adminRoutes.Delete("/news/:uuid", newsHandler.Delete)

	// activities
	protected.Get("/activity-categories", activityHandler.Categories)
	protected.Get("/activities", activityHandler.List)
	protected.Get("/activities/:uuid", activityHandler.Get)
	protected.Post("/activities/:uuid/submit", submissionHandler.Submit)
	protected.Get("/activity-submissions/me", submissionHandler.Mine)

	// sites
	protected.Get("/sites", sitesHandler.List)

	// contacts

	// points
	protected.Get("/points", pointsHandler.Get)
	protected.Post("/points/spend", pointsHandler.Spend)
	protected.Get("/rewards", rewardsHandler.List)
	protected.Post("/rewards/:uuid/redeem", rewardsHandler.Redeem)
	protected.Get("/reward-redemptions/me", rewardsHandler.MyRedemptions)

	// attendance
	protected.Post("/attendance/check-in", attendanceHandler.CheckIn)
	protected.Get("/attendance/me", attendanceHandler.GetMyHistory)
	protected.Get("/attendance/summary", attendanceHandler.GetMySummary)
	adminRoutes.Get("/attendance", attendanceHandler.AdminReport)
	adminRoutes.Get("/attendance/summary", attendanceHandler.AdminSummary)

	// streaks
	protected.Get("/streaks", streakHandler.List)
	protected.Post("/streaks/check-in", streakHandler.CheckIn)

	// news
	protected.Get("/news", newsHandler.List)
	adminRoutes.Post("/news/:uuid/image", newsHandler.UploadImage)
}
