package model

type Article struct {
	ID      int    `json:"id" gorm:"type:int(11);primaryKey;auto_increment;comment:ID"`
	Title   string `json:"title" gorm:"type:varchar(32);not null;default:'';comment:标题"`
	Created uint32 `json:"created" gorm:"type:int(11);not null;autoCreateTime;comment:创建时间"`
	Updated uint32 `json:"updated" gorm:"type:int(11);not null;autoUpdateTime;comment:更新时间"`
}
