package iofxml_test

import (
	"fmt"
	"strings"

	"github.com/MetsaApp/iofxml"
)

func ExampleRead() {
	body := `<ResultList xmlns="http://www.orienteering.org/datastandard/3.0" iofVersion="3.0">
  <Event><Name>Example Open</Name></Event>
  <ClassResult>
    <Class><Name>H21</Name></Class>
    <PersonResult>
      <Person><Name><Family>Pine</Family><Given>Paul</Given></Name></Person>
      <Result><Time>2112.4</Time><Position>1</Position><Status>OK</Status></Result>
    </PersonResult>
  </ClassResult>
</ResultList>`
	doc, err := iofxml.Read(strings.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, c := range doc.ResultList.ClassResults {
		for _, p := range c.PersonResults {
			r := p.Results[0]
			fmt.Println(c.Class.Name, p.Person.Name.Full(), r.Status, *r.Time)
		}
	}
	fmt.Println(doc.ResultList.EffectiveStatus())
	// Output:
	// H21 Paul Pine OK 2112.4
	// Complete
}

func ExampleMarshal() {
	list := &iofxml.EntryList{
		Event: iofxml.Event{Name: "Example Open"},
		PersonEntries: []iofxml.PersonEntry{{
			Person:  iofxml.Person{Name: iofxml.PersonName{Family: "Alder", Given: "Ada"}},
			Classes: []iofxml.Class{{ID: iofxml.ID{Value: "1"}, Name: "D21"}},
		}},
	}
	list.CreateTime = "2026-09-01T12:00:00Z"
	list.Creator = "Example Timing 1.0"
	out, err := iofxml.Marshal(list)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(string(out))
	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <EntryList xmlns="http://www.orienteering.org/datastandard/3.0" iofVersion="3.0" createTime="2026-09-01T12:00:00Z" creator="Example Timing 1.0">
	//   <Event>
	//     <Name>Example Open</Name>
	//   </Event>
	//   <PersonEntry>
	//     <Person>
	//       <Name>
	//         <Family>Alder</Family>
	//         <Given>Ada</Given>
	//       </Name>
	//     </Person>
	//     <Class>
	//       <Id>1</Id>
	//       <Name>D21</Name>
	//     </Class>
	//   </PersonEntry>
	// </EntryList>
}
