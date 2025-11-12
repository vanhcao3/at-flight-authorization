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
  "airport": "HAN",
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

## Update Flight Authorization Proposal

Replace `{proposal_id}` with the identifier of the proposal you want to update. Provide the complete proposal, including nested operator, drones, flight areas, and pilot details.

```bash
curl -X PUT https://airtransys.site:9443/flight-authorization/flight-authorization-proposals/{proposal_id} \
  -H 'Content-Type: application/json' \
  -d @proposal-update.json
```

`proposal-update.json`:

```json
{
  "name": "Survey Flight 42 - Revised Plan",
  "airport": "HAN",
  "flight_purpose": "Infrastructure inspection",
  "operator": {
    "business_name": "Aero Dynamics Co.",
    "tax_identification_number": "123456789",
    "address": "12 Tran Hung Dao, Hoan Kiem, Hanoi",
    "nationality": "VNM",
    "phone_number": "+84-24-12345678",
    "fax": "+84-24-98765432",
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
    },
    {
      "drone_specification": {
        "maximum_take_off_weight": 12,
        "dimension": {
          "length": 1,
          "width": 2,
          "height": 1
        },
        "engine_type": "Electric",
        "operating_frequency": "5.8GHz",
        "operating_method": "Remote"
      },
      "drone_registration": {
        "drone_type": "Inspector Lite",
        "factory_number": "IL-014",
        "registration_number": "VN-IL-014",
        "registration_date": "2024-06-18"
      }
    }
  ],
  "flight_area": [
    {
      "place": "Noi Bai East Sector",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "altitude": "180m",
      "polygon": [
        {
          "latitude": 21.2191,
          "longitude": 105.8054
        },
        {
          "latitude": 21.2214,
          "longitude": 105.8106
        },
        {
          "latitude": 21.2167,
          "longitude": 105.8128
        }
      ]
    }
  ],
  "operating_duration": {
    "duration": 2,
    "from_day": "2025-11-25T07:00:00Z",
    "to_day": "2025-11-26T16:30:00Z"
  },
  "pilot": {
    "name": "Nguyen Van B",
    "birthday": "1988-08-21",
    "identification_number": "987654321",
    "phone_number": "+84-913579246",
    "pilot_license": {
      "license_number": "LIC-2024-12",
      "license_provision_date": "2024-02-15"
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
  "flight_authorization_proposal_id": "9a671807-d28c-4f2c-8d67-d3a5cc086a86",
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

## Update Flight Authorization Approval

Replace `{approval_id}` with the identifier of the approval you want to update.

```bash
curl -X PUT https://airtransys.site:9443/flight-authorization/flight-authorization-approvals/{approval_id} \
  -H 'Content-Type: application/json' \
  -d @approval-update.json
```

`approval-update.json`:

```json
{
  "name": "Approval for Survey Flight 42 - Revised",
  "flight_authorization_proposal_id": "9a671807-d28c-4f2c-8d67-d3a5cc086a86",
  "flight_negotiation_authorities": [
    "Civil Aviation Authority",
    "Air Defense Liaison"
  ],
  "authorized_operating_duration": {
    "duration": 2,
    "from_day": "2025-11-25T07:00:00Z",
    "to_day": "2025-11-26T16:30:00Z"
  },
  "authorized_flight_area": [
    {
      "place": "Noi Bai East Sector",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "altitude": "180m",
      "polygon": [
        {
          "latitude": 21.2191,
          "longitude": 105.8054
        },
        {
          "latitude": 21.2214,
          "longitude": 105.8106
        },
        {
          "latitude": 21.2167,
          "longitude": 105.8128
        }
      ]
    }
  ],
  "flight_parameter": {
    "altitude": "180m",
    "radius": "6km",
    "take_off_and_landing_area": {
      "place": "Noi Bai Hangar 4",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "latitude": 21.2139,
      "longitude": 105.8031
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
  "flight_authorization_approval_id": "d9310e3d-1c43-4f3b-8e48-7f7ea306a8e4",
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

## Update Flight Notification

Replace `{notification_id}` with the identifier of the notification you want to update.

```bash
curl -X PUT https://airtransys.site:9443/flight-authorization/flight-notifications/{notification_id} \
  -H 'Content-Type: application/json' \
  -d @notification-update.json
```

`notification-update.json`:

```json
{
  "name": "Notification for Survey Flight 42 - Revised",
  "flight_authorization_approval_id": "d9310e3d-1c43-4f3b-8e48-7f7ea306a8e4",
  "intended_operating_duration": {
    "duration": 1,
    "from_day": "2025-11-25T09:00:00Z",
    "to_day": "2025-11-25T16:30:00Z"
  },
  "intended_flight_area": [
    {
      "place": "Noi Bai East Sector",
      "commune": "Phu Cuong",
      "province": "Hanoi",
      "altitude": "180m",
      "polygon": [
        {
          "latitude": 21.2191,
          "longitude": 105.8054
        },
        {
          "latitude": 21.2214,
          "longitude": 105.8106
        },
        {
          "latitude": 21.2167,
          "longitude": 105.8128
        }
      ]
    }
  ]
}
```
