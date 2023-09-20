package kvengine

type Status string

const (
	//BST status
	nilBST           = "nilBST"
	SetSuccess  = "SetSuccess"
	SetFailed = "SetFailed"
	DeleteSuccess    = "DeleteSuccess"
	DeleteDeletedVal = "DeleteDeletedVal"
	DeleteNotFound   = "DeleteNotFound"

	//WAL status
	WriteFailed = "WriteFailed"
	WriteSuccess = "WriteSuccess"


	//sstable search status
	SStableSearchDeleted = "SStableSearchDeleted"
	SStableSearchNone = "SStableSearchNone"
	SStableSearchSuccess = "SStableSearchSuccess"
)
