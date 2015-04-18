package configserver

//mock configserver for test
type mock_cfs struct {
   bucketTbl func() BucketTbl
   bucketTblRef func() *BucketTbl
   updateBucketTbl func(*BucketTbl)error
   getProperty func(string)string
   setProperty func(string, string)error
   localCsId func()uint64
}

func (ms *mock_cfs) BucketTbl() BucketTbl {
   return ms.bucketTbl()
}

func (ms *mock_cfs) BucketTblRef() *BucketTbl {
   return ms.bucketTblRef()
}

func (ms *mock_cfs) UpdateBucketTbl(tbl *BucketTbl) error {
   return ms.updateBucketTbl(tbl)
}

func (ms *mock_cfs) GetProperty(key string) string {
   return ms.getProperty(key)
}

func (ms *mock_cfs) SetProperty(key string, value string) error {
   return ms.SetProperty(key, value)
}

func (ms *mock_cfs) LocalCsId() uint64 {
   return ms.localCsId()
}

