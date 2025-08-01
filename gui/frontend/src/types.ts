// Common types used across the application

export interface Person {
  id: string;
  lng: number;
  lat: number;
  raw_xml: string;
}

export interface Activity {
  id: string;
  type: string;
  location: [number, number];
  startTime: string;
  endTime: string;
}

export interface Plan {
  type: 'weekday' | 'weekend' | 'holiday';
  activities: Activity[];
}