package dto

type DailyActivityStat struct {
	Date            string `json:"date"`
	TotalActivities int64  `json:"total_activities"`
}

type IssueVelocityStat struct {
	Date            string `json:"date"`
	IssuesCreated   int64  `json:"issues_created"`
	IssuesCompleted int64  `json:"issues_completed"`
}

type ActivityTypeBreakdown struct {
	ActivityType string `json:"activity_type"`
	Count        int64  `json:"count"`
}

type MemberActivityStat struct {
	UserID               int64  `json:"user_id"`
	Fullname             string `json:"fullname"`
	Username             string `json:"username"`
	Email                string `json:"email"`
	IssueActionsCount    int64  `json:"issue_actions_count"`
	CompletedIssuesCount int64  `json:"completed_issues_count"`
}

type WorkspaceAnalyticsResponse struct {
	DailyActivities   []DailyActivityStat     `json:"daily_activities"`
	IssueVelocity     []IssueVelocityStat     `json:"issue_velocity"`
	ActivityBreakdown []ActivityTypeBreakdown `json:"activity_breakdown"`
	MemberLeaderboard []MemberActivityStat    `json:"member_leaderboard"`
}

type ProjectAnalyticsResponse struct {
	DailyActivities   []DailyActivityStat     `json:"daily_activities"`
	IssueVelocity     []IssueVelocityStat     `json:"issue_velocity"`
	ActivityBreakdown []ActivityTypeBreakdown `json:"activity_breakdown"`
	MemberLeaderboard []MemberActivityStat    `json:"member_leaderboard"`
}
