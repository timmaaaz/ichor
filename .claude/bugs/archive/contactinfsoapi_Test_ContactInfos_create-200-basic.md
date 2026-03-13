# Test Failure: Test_ContactInfos/create-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/core/contactinfsoapi`
- **Duration**: 0.02s

## Failure Output

```
    apitest.go:73: DIFF
    apitest.go:74:   &contactinfosapp.ContactInfos{
          	... // 6 identical fields
          	StreetID:            "a2fc313a-455b-4fd6-afe6-da3454509354",
          	DeliveryAddressID:   "895e02b9-477b-4e6c-a25b-e710a2be3ab2",
        - 	AvailableHoursStart: "08:00:00",
        + 	AvailableHoursStart: "8:00:00",
          	AvailableHoursEnd:   "17:00:00",
          	TimezoneID:          "b443003f-73dc-4e6d-b0bf-6749215eec1c",
          	... // 2 identical fields
          }
    apitest.go:75: GOT
    apitest.go:76: &contactinfosapp.ContactInfos{ID:"0d542fc5-2ae7-46ef-a9cc-b0f8d628900e", FirstName:"John", LastName:"Doe", EmailAddress:"johndoe@example.com", PrimaryPhone:"+1234567890", SecondaryPhone:"", StreetID:"a2fc313a-455b-4fd6-afe6-da3454509354", DeliveryAddressID:"895e02b9-477b-4e6c-a25b-e710a2be3ab2", AvailableHoursStart:"08:00:00", AvailableHoursEnd:"17:00:00", TimezoneID:"b443003f-73dc-4e6d-b0bf-6749215eec1c", PreferredContactType:"email", Notes:""}
    apitest.go:77: EXP
    apitest.go:78: &contactinfosapp.ContactInfos{ID:"0d542fc5-2ae7-46ef-a9cc-b0f8d628900e", FirstName:"John", LastName:"Doe", EmailAddress:"johndoe@example.com", PrimaryPhone:"+1234567890", SecondaryPhone:"", StreetID:"a2fc313a-455b-4fd6-afe6-da3454509354", DeliveryAddressID:"895e02b9-477b-4e6c-a25b-e710a2be3ab2", AvailableHoursStart:"8:00:00", AvailableHoursEnd:"17:00:00", TimezoneID:"b443003f-73dc-4e6d-b0bf-6749215eec1c", PreferredContactType:"email", Notes:""}
    apitest.go:79: Should get the expected response
--- FAIL: Test_ContactInfos/create-200-basic (0.02s)
```

## Fix

- **File**: `api/cmd/services/ichor/tests/core/contactinfsoapi/create_test.go:40`
- **Classification**: test bug
- **Change**: Updated `ExpResp.AvailableHoursStart` from `"8:00:00"` to `"08:00:00"` — PostgreSQL returns time values in canonical HH:MM:SS form with leading zeros
- **Verified**: `go test -v -run Test_ContactInfos/create-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/core/contactinfsoapi` ✓
