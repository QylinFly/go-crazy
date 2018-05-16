package Routers

import (
	"fmt"
	"go.uber.org/zap"
	. "github.com/xoxo/crm-x/util"
	"github.com/xoxo/crm-x/util/logger"
	Gin "github.com/gin-gonic/gin"
	"github.com/xoxo/crm-x/app/Services/Elastic/Models"
	"github.com/olivere/elastic"
	"github.com/xoxo/crm-x/app/Services/Elastic/Utils"
	"github.com/golang/net/context"
	// "github.com/xoxo/crm-x/app/Services/Elastic/Entity"
	"encoding/json"
)

var Index = "www"

func GetHandler(c *Gin.Context)  {

// if 1==0 {
// 	client2 := Models.GetElasticCon(Utils.ElasticUrl());
// 	{
// 		// Search with a term query
// 		 termQuery := elastic.NewTermQuery("_id", "1")
// 		searchResult, err := client2.Search().
// 			Index("www").   // search in index "twitter"
// 			Query(termQuery).  // specify the query
// 			Type("external").
// 			Sort("_id", true). // sort by "user" field, ascending
// 			From(0).Size(10).   // take documents 0-9
// 			Pretty(true).       // pretty print request and response JSON
// 			Do(context.TODO())                // execute
// 		if err != nil {
// 			// Handle error
// 			panic(err)
// 		}

// 		// searchResult is of type SearchResult and returns hits, suggestions,
// 		// and all kinds of other information from Elasticsearch.
// 		log.Printf("Query took %d milliseconds\n", searchResult.TotalHits())

// 		    // Iterate through results
// 			for _, hit := range searchResult.Hits.Hits {
// 				// hit.Index contains the name of the index
// 				j, err4 := json.Marshal(&hit.Source)
// 				if err4 != nil {
// 					panic(err4)
// 				}
// 				log.Printf(string(j))  
// 			}

// 	}
// 	// Get tweet with specified ID
	
// 	res, err2 := client2.Get().Index("www").Id("1").Do(context.TODO());
// 	if err2 != nil {
// 		log.Printf("expected Found = true; got %v", err2)
// 	}
// 	 if res.Found == true {

// 		log.Printf("expected Found = true; got %v", res.Source)
// 	 }

// 	 j, err4 := json.Marshal(&res.Source)
//     if err4 != nil {
//         panic(err4)
//     }
//     log.Printf(string(j))  

// 	return
// }

	client := Models.GetElasticCon(Utils.ElasticUrl());
	dat := Utils.BodyToJson(c.Request)
	eType := dat["type"].(string)
	Index := dat["index"].(string)
	query_type := dat["query_type"].(string)
	child_type := dat["child_type"].(string)
	start_index := int(dat["start_index"].(float64))
	array_of_json := dat["query_json"].([]interface{});
	size := int(dat["size"].(float64))
	sorting,err1 := dat["sort"].(map[string]interface{});
	if err1 != true {
		panic(err1)
	}
	var  fieldName string
	var  sortType bool
	for i:= range sorting{
		if(i=="field"){
			fieldName =  sorting[i].(string)
		}else if(i=="asc"){
			sortType = true;
		}
	}
	bq := elastic.NewBoolQuery()
	if query_type=="parent" {
		datRecord := array_of_json[0]
		res := datRecord.(map[string]interface{})
		key := res["key"].(string)
		value := res["value"].(string)

		matchChildQuery := elastic.NewHasChildQuery(child_type, elastic.NewMatchQuery(key , value)).
			InnerHit(elastic.NewInnerHit().Name("age"))
		bq = bq.Must(elastic.NewMatchAllQuery())
		bq = bq.Filter(matchChildQuery)

	}else {
		//newQ := elastic.NewBoolQuery()
		for i := 0; i < len(array_of_json); i++ {
			datRecord := array_of_json[i]
			res := datRecord.(map[string]interface{})
			qType := res["query_type"].(string)
			matchQueryType := res["match"].(string)
			key := res["key"].(string)
			value := res["value"].(interface{})
			//switch vv := value.(type) {
			//case string:
			//
			//case int:
			//
			//case []interface{}:
			//	for i, u := range vv {
			//		fmt.Println(i, u)
			//	}
			//default:
			//	fmt.Println(k, "is of a type I don't know how to handle")
			//}
			var matchType  *elastic.MatchQuery
			var termQuery *elastic.TermQuery
			var rangeQuery *elastic.RangeQuery
			match := 0
			switch matchQueryType {
			case "text" :
				value := res["value"].(string)
				matchType = elastic.NewMatchQuery(key,value)
				break;
			case "keyword" :
				match = 1
				termQuery = elastic.NewTermQuery(key,value)
				break;
			case "range" :
				match = 2
				rangeQuery = elastic.NewRangeQuery(key)
				valueRange := value.(map[string]interface{})

				for i := range valueRange{
					switch i {
					case "gte" :
						rangeQuery =  rangeQuery.Gte(valueRange[i])
						break
					case "gt" :
						rangeQuery =  rangeQuery.Gt(valueRange[i])
						break
					case "lte" :
						rangeQuery =  rangeQuery.Lte(valueRange[i])
						break
					case "lt" :
						rangeQuery =  rangeQuery.Lt(valueRange[i])
						break
					}
				}
				break;
			}
			switch qType {
			case "must" :
				if match == 0 {
					bq = bq.Must(matchType)
				}else {
					bq = bq.Must(termQuery)
				}
				break;
			case "filter":
				if match ==0 {
					bq = bq.Filter(matchType)
				}else {
					bq = bq.Filter(termQuery)
				}
				break;
			case "must_not":
				if match ==0 {
					bq = bq.MustNot(matchType)
				}else {
					bq = bq.MustNot(termQuery)
				}
				break;
			case "should":
				if match==0 {
					bq =  bq.Should(matchType)
				}else {
					bq =  bq.Should(termQuery)
				}
				break;
			}
			//newQ = newQ.Should(matchType)
		}
		//bq.Filter(newQ)
	}

	var searchResult *elastic.SearchResult
	eQuery := client.Search().
		Index(Index).
		Type(eType).
		Query(bq).
		From(start_index).
		Size(size).
		TrackScores(true). // score
		Version(true)	// version
	if(fieldName != ""){
		eQuery = eQuery.Sort(fieldName,sortType)
	}
	searchResult,err := eQuery.Pretty(true).Do(context.Background())
	if err!= nil {
		panic(err)
	}
	hits := searchResult.Hits.Hits

	datArray := make([]map[string]interface{},len(hits))
	
	for i := 0;i < len(hits) ; i++ {
		var dat1 map[string]interface{}
		hit := searchResult.Hits.Hits[i]
		if err := json.Unmarshal(*hit.Source,&dat1); err != nil {
			panic(err)
		}

		logger.Debug("Elastic Get",zap.String("Hits", fmt.Sprint(dat1))) //fmt.Sprint(dat1)
		
		dat1["_id"]=hit.Id
		dat1["_type"]=hit.Type
		dat1["_version"]=hit.Version
		dat1["_score"]=hit.Score
		
		datArray[i] = dat1;
	}

	Api_response(c,Gin.H{"data_source":datArray,"status" :true,"length" :len(hits)})
}
