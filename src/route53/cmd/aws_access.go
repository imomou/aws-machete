package cmd

import (
	//"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type route53Management interface {
	GetAllRecordSets() (records []*route53.ResourceRecordSet, err error)
}

type route53Manager struct {
	route53client route53iface.Route53API
}

func newRoute53client() route53Management {
	var sess = session.Must(session.NewSession(&aws.Config{}))

	var result route53Management = &route53Manager{
		route53client: route53.New(sess)}

	return result
}

func (client *route53Manager) GetAllRecordSets() (records []*route53.ResourceRecordSet, err error) {

	hostedZones, err := client.route53client.ListHostedZones(&route53.ListHostedZonesInput{})

	if err != nil {
		return nil, err
	}

	recordSets := make([]*route53.ResourceRecordSet, 0)
	for _, zone := range hostedZones.HostedZones {

		results, err := client.route53client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
			HostedZoneId: zone.Id})

		recordSets = append(recordSets, results.ResourceRecordSets...)

		if err != nil {
			return []*route53.ResourceRecordSet{}, err
		}

		for *results.IsTruncated {

			results, err = client.route53client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
				StartRecordName: results.NextRecordName,
				StartRecordType: results.NextRecordType,
				HostedZoneId:    zone.Id})

			recordSets = append(recordSets, results.ResourceRecordSets...)
		}
	}

	return recordSets, nil
}
