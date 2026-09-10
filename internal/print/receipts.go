package print

func CommonBlocks() (dateTime, address, transactionID, stornID, voucherCode, phoneNumber *Template) {
	b := NewBuilder()
	dateTime = b.Line().Param("dts").Template()

	b = NewBuilder()
	address = b.Line().Param("mn").Text(",").Param("pa").Template()

	b = NewBuilder()
	transactionID = b.Line().Text("ID transakcije").Line().Param("tid").Template()

	b = NewBuilder()
	stornID = b.Line().Text("ID stornirane transakcije").Line().Param("stid").Template()

	b = NewBuilder()
	voucherCode = b.Line().Text("14-znamenkasti kod:").Line().Param("vc").Template()

	b = NewBuilder()
	phoneNumber = b.Line().Text("Broj telefona: ").Param("phN").Template()
	return
}

func PromoMessages() *Template {
	b := NewBuilder()
	return b.
		Line().Custom(func(p Params) string { return StripPromoPrefix(p["pm1"]) }).
		Line().Custom(func(p Params) string { return StripPromoPrefix(p["pm2"]) }).
		Line().Custom(func(p Params) string { return StripPromoPrefix(p["pm3"]) }).
		Template()
}

func Receipt(brand, kind string, extra ...*Template) *Template {
	dateTime, address, tid, storn, voucher, phone := CommonBlocks()
	b := NewBuilder()
	b.Add(dateTime).Add(address)
	b.Line().Text(brand)
	switch kind {
	case "charge":
		b.Line().Text(brand + " storniranje dopune")
		b.Line().Param("amt").Text(" KM")
		b.Add(storn).Add(tid)
	case "last":
		b.Line().Text("Kopija posljednje dopune")
		b.Line().Param("amt").Text(" KM")
		b.Add(tid)
		b.Add(phone)
	case "voucher":
		b.Line().Param("amt").Text(" KM")
		b.Add(tid)
		b.Add(voucher)
	default:
		b.Line().Param("amt").Text(" KM")
		b.Add(tid)
		b.Add(phone)
	}
	for _, e := range extra {
		if e != nil {
			b.Add(e)
		}
	}
	b.Line()
	return b.Template()
}

func SumReport() *Template {
	b := NewBuilder()
	return b.
		Line().Text("Sumarni izvještaj").
		Line().Param("mo").
		Line().Text("Za datum ").Param("rh").
		Line().Param("fromTime").Text("-").Param("toTime").
		Line().Param("mn").Text(",").Param("pa").
		Report(3).ReportColumn().ReportColumn().ReportColumn().
		LineAlign(AlignLeft).Text("Ukupan iznos:").Space(1).Param("sum3").Space(1).Text("KM").
		LineAlign(AlignLeft).Text("Štampano: ").Param("rf").
		Line().
		Template()
}

func DetailReport(title string) *Template {
	b := NewBuilder()
	return b.
		Line().Text(title).
		Line().Param("mo").
		Line().Text("Za datum ").Param("rh").
		Line().Param("fromTime").Text("-").Param("toTime").
		Line().Param("mn").Text(",").Param("pa").
		Report(3).ReportColumn().ReportColumn().ReportColumn().
		LineAlign(AlignLeft).Text("Broj transakcija:").Space(1).Param("sum1").
		LineAlign(AlignLeft).Text("Ukupan iznos:").Space(1).Param("sum2").Space(1).Text("KM").
		LineAlign(AlignLeft).Text("Štampano: ").Param("rf").
		Line().
		Template()
}

func WeekReport() *Template {
	b := NewBuilder()
	return b.
		Line().Param("dnreport").
		Line().Param("da").
		Line().Text("Tel: ").Param("dt").
		Line().Param("bn1").Line().Param("ba1").
		Line().Param("bn2").Line().Param("ba2").
		Line().Param("bn3").Line().Param("ba3").
		Line().Param("bn4").Line().Param("ba4").
		Line().Text("Za: ").Param("pn").
		Line().Param("pa").
		Line().Text("Izvještaj za period").
		Line().Param("f").Param("t").
		Report(2).ReportColumn().ReportColumn().
		LineAlign(AlignLeft).Text("Ukupan iznos:").Space(1).Param("sum2").Space(1).Text("KM").
		LineAlign(AlignLeft).Text("Štampano: ").Param("rf").
		Line().
		Template()
}
