# Flight Authorization Payload Examples

Use `Content-Type: application/json` with every request. The service generates identifiers for all nested objects, so the payloads do not need `id` or foreign key fields.

## Create Flight Authorization Proposal

```bash
curl -X POST https://airtransys.site:9443/flight-authorization/flight-authorization-proposals \
  -H 'Content-Type: application/json' \
  -d @proposal.json
```

`proposal.json`:

```json
{
  "name": "Survey Flight 42",
  "take_off_and_landing_area": {
      "place": "Noi Bai Hangar 3",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "latitude": 21.2142,
      "longitude": 105.8025
  },
  "flight_purpose": "Aerial survey",
  "file_paths": ["s3://bucket/plans/survey-flight-42.pdf"],
  "operator": {
    "business_name": "Aero Dynamics Co.",
    "tax_identification_number": "123456789",
    "address": "12 Tran Hung Dao, Hoan Kiem, Hanoi",
    "nationality": "VNM",
    "phone_number": "+84-24-12345678",
    "fax": "+84-24-12345679",
    "email": "ops@aerodynamics.vn"
  },
  "drones": [
    {
      "drone_specification": {
        "maximum_take_off_weight": 18,
        "dimension": {
          "length": 2,
          "width": 3,
          "height": 1
        },
        "engine_type": "Hybrid",
        "operating_frequency": "2.4GHz",
        "operating_method": "Remote"
      },
      "drone_registration": {
        "drone_type": "Surveyor X",
        "factory_number": "SX-001",
        "registration_number": "VN-SX-001",
        "registration_date": "2024-03-10"
      }
    }
  ],
  "flight_area": [
    {
      "place": "Noi Bai East Sector",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "altitude": "150m",
      "polygon": [
        {
          "latitude": 21.2185,
          "longitude": 105.8041
        },
        {
          "latitude": 21.2201,
          "longitude": 105.8093
        },
        {
          "latitude": 21.2154,
          "longitude": 105.8119
        }
      ]
    }
  ],
  "operating_duration": {
    "duration": 3,
    "from_day": "2025-11-20T08:00:00Z",
    "to_day": "2025-11-22T17:00:00Z"
  },
  "pilot": {
    "name": "Nguyen Van A",
    "birthday": "1990-04-12",
    "identification_number": "123456789",
    "phone_number": "+84-912345678",
    "pilot_license": {
      "license_number": "LIC-2023-88",
      "license_provision_date": "2023-05-01"
    }
  }
}
```

## Create Flight Authorization Approval

```bash
curl -X POST https://airtransys.site:9443/flight-authorization/flight-authorization-approvals \
  -H 'Content-Type: application/json' \
  -d @approval.json
```

`approval.json`:

```json
{
  "name": "Approval for Survey Flight 42",
  "flight_authorization_proposal_id": "4e5e6832-a1bb-4ceb-9a70-17a61ba3ac62",
  "flight_negotiation_authorities": ["Civil Aviation Authority"],
  "authorized_operating_duration": {
    "duration": 3,
    "from_day": "2025-11-20T08:00:00Z",
    "to_day": "2025-11-22T17:00:00Z"
  },
  "authorized_flight_area": [
    {
      "place": "Noi Bai East Sector",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "altitude": "150m",
      "polygon": [
        {
          "latitude": 21.2185,
          "longitude": 105.8041
        },
        {
          "latitude": 21.2201,
          "longitude": 105.8093
        },
        {
          "latitude": 21.2154,
          "longitude": 105.8119
        }
      ]
    }
  ],
  "flight_parameter": {
    "altitude": "150m",
    "radius": "5km",
    "take_off_and_landing_area": {
      "place": "Noi Bai Hangar 3",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "latitude": 21.2142,
      "longitude": 105.8025
    }
  }
}
```

## Create Flight Notification

```bash
curl -X POST https://airtransys.site:9443/flight-authorization/flight-notifications \
  -H 'Content-Type: application/json' \
  -d @notification.json
```

`notification.json`:

```json
{
  "name": "Notification for Survey Flight 42",
  "flight_authorization_approval_id": "8410246e-c145-49b3-bd38-bd3decc717b5",
  "intended_operating_duration": {
    "duration": 1,
    "from_day": "2025-11-20T08:00:00Z",
    "to_day": "2025-11-20T17:00:00Z"
  },
  "intended_flight_area": [
    {
      "place": "Noi Bai East Sector",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "altitude": "150m",
      "polygon": [
        {
          "latitude": 21.2185,
          "longitude": 105.8041
        },
        {
          "latitude": 21.2201,
          "longitude": 105.8093
        },
        {
          "latitude": 21.2154,
          "longitude": 105.8119
        }
      ]
    }
  ]
}
```
