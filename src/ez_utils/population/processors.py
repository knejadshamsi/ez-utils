from bs4 import BeautifulSoup
import pandas as pd
import numpy as np
import os
from pathlib import Path
from typing import List, Dict, Tuple, Union
from collections import defaultdict
from .models import Activity, Person, PopulationData, ScaledPopulation
from .utils import create_point_from_coordinates, parse_time, get_total_population
from .base_processor import BaseProcessor

def validate_scales(current_scale: float, requested_scales: Union[List[float], str, None] = None) -> List[float]:
    if current_scale < 10:
        raise ValueError("Current population is less than 10% of total population. Cannot generate scaled versions.")
        
    if requested_scales is None:
        scales = list(range(1, 11))
    else:
        # Handle comma-separated string input
        if isinstance(requested_scales, str):
            try:
                requested_scales = [float(s.strip()) for s in requested_scales.split(',')]
            except ValueError:
                raise ValueError("Invalid scale format. Use comma-separated numbers between 1 and 10")
        
        scales = [float(s) for s in requested_scales if 1 <= float(s) <= 10]
        if not scales:
            raise ValueError("No valid scales provided. Scales must be between 1 and 10.")
    
    valid_scales = [s for s in scales if s <= current_scale]
    if not valid_scales:
        raise ValueError(f"No valid scales below current scale ({current_scale}%). This module only supports scaling down.")
    return sorted(valid_scales)

def validate_population_xml(soup: BeautifulSoup) -> None:
    if not soup.find('population'):
        raise ValueError("Invalid population XML: Missing 'population' root element")
    if not soup.find_all('person'):
        raise ValueError("Invalid population XML: No persons found")
    for person in soup.find_all('person'):
        if not person.find('plan'):
            raise ValueError(f"Invalid person: Missing plan for person {person.get('id', 'unknown')}")
        if not person.find('activity'):
            raise ValueError(f"Invalid person: No activities found for person {person.get('id', 'unknown')}")
        first_activity = person.find('activity')
        if 'x' not in first_activity.attrs or 'y' not in first_activity.attrs:
            raise ValueError(f"Invalid person: First activity missing coordinates for person {person.get('id', 'unknown')}")

def calculate_location_density(activities_df: pd.DataFrame) -> Dict[tuple, int]:
    location_counts = defaultdict(int)
    first_activities = activities_df[activities_df['activity_order'] == 1]
    
    for _, activity in first_activities.iterrows():
        if activity['coordinates']:
            location = (activity['coordinates'].x, activity['coordinates'].y)
            location_counts[location] += 1
    
    if not location_counts:
        raise ValueError("No valid locations found in first activities")
    
    return location_counts

def select_persons_by_density(persons: List[Person], activities_df: pd.DataFrame, 
                            target_count: int) -> List[Person]:
    if target_count <= 0:
        raise ValueError("Target count must be positive")
    if target_count > len(persons):
        raise ValueError("Target count cannot be greater than current population")
        
    location_density = calculate_location_density(activities_df)
    total_population = len(persons)
    scale_factor = target_count / total_population
    
    selected_persons = []
    location_selections = defaultdict(int)
    
    persons_by_location = defaultdict(list)
    for person in persons:
        first_activity = activities_df[
            (activities_df['id'] == person.id) & 
            (activities_df['activity_order'] == 1)
        ].iloc[0]
        
        if first_activity['coordinates']:
            location = (first_activity['coordinates'].x, first_activity['coordinates'].y)
            persons_by_location[location].append(person)
    
    for location, count in location_density.items():
        target_location_count = int(count * scale_factor)
        available_persons = persons_by_location[location]
        
        if available_persons:
            selected = available_persons[:target_location_count]
            selected_persons.extend(selected)
            location_selections[location] = len(selected)
    
    if not selected_persons:
        raise ValueError("No persons could be selected based on density criteria")
        
    return selected_persons[:target_count]

def process_activities(person_soup: BeautifulSoup) -> Tuple[List[Activity], pd.Series]:
    activities = []
    activity_data = {}
    
    person_id = person_soup['id']
    plan = person_soup.find('plan')
    if not plan:
        raise ValueError(f"No plan found for person {person_id}")
    
    for i, activity in enumerate(plan.find_all('activity'), start=1):
        if 'type' not in activity.attrs:
            raise ValueError(f"Activity missing type attribute for person {person_id}")
            
        if 'x' in activity.attrs and 'y' in activity.attrs:
            try:
                coordinates = create_point_from_coordinates(
                    float(activity['x']), float(activity['y'])
                )
            except ValueError:
                raise ValueError(f"Invalid coordinates for person {person_id}, activity {i}")
        else:
            coordinates = None
            
        activity_obj = Activity(
            id=person_id,
            activity_order=i,
            activity_type=activity.get('type'),
            facility=activity.get('facility'),
            start_time=parse_time(activity.get('start_time')),
            end_time=parse_time(activity.get('end_time')),
            coordinates=coordinates
        )
        activities.append(activity_obj)
        
        activity_data.update({
            'id': person_id,
            'activity_order': i,
            'activity_type': activity_obj.activity_type,
            'facility': activity_obj.facility,
            'start_time': activity_obj.start_time,
            'end_time': activity_obj.end_time,
            'coordinates': activity_obj.coordinates
        })
    
    if not activities:
        raise ValueError(f"No activities found for person {person_id}")
    
    return activities, pd.Series(activity_data)

def process_population(population_soup: BeautifulSoup) -> PopulationData:
    validate_population_xml(population_soup)
    persons = []
    activities_data = []
    
    for person in population_soup.find_all('person'):
        activities, activity_data = process_activities(person)
        persons.append(Person(id=person['id'], activities=activities))
        activities_data.append(activity_data)
    
    activities_df = pd.DataFrame(activities_data)
    
    try:
        total_population = get_total_population()
        if total_population <= 0:
            raise ValueError("TOTAL_POPULATION must be positive")
    except (KeyError, ValueError) as e:
        raise ValueError("TOTAL_POPULATION environment variable not set or invalid") from e
        
    current_scale = (len(persons) / total_population) * 100
    if current_scale <= 0:
        raise ValueError("Current population scale cannot be zero or negative")
    
    return PopulationData(
        persons=persons,
        activities_df=activities_df,
        total_population=total_population,
        current_scale=current_scale
    )

def create_scaled_population(population_data: PopulationData, target_scale: float) -> ScaledPopulation:
    if target_scale <= 0:
        raise ValueError("Target scale must be positive")
    if target_scale > population_data.current_scale:
        raise ValueError(f"Target scale ({target_scale}%) cannot be greater than current scale ({population_data.current_scale}%)")
    
    scale_ratio = target_scale / population_data.current_scale
    target_count = int(len(population_data.persons) * scale_ratio)
    
    selected_persons = select_persons_by_density(
        population_data.persons,
        population_data.activities_df,
        target_count
    )
    
    selected_activities = population_data.activities_df[
        population_data.activities_df['id'].isin([p.id for p in selected_persons])
    ]
    
    return ScaledPopulation(
        scale=target_scale,
        persons=selected_persons,
        activities_df=selected_activities
    )

def population_to_xml(population: ScaledPopulation) -> str:
    soup = BeautifulSoup('<population></population>', 'lxml-xml')
    root = soup.find('population')
    
    for person in population.persons:
        person_tag = soup.new_tag('person')
        person_tag['id'] = person.id
        
        plan = soup.new_tag('plan')
        person_tag.append(plan)
        
        person_activities = population.activities_df[
            population.activities_df['id'] == person.id
        ].sort_values('activity_order')
        
        for _, activity in person_activities.iterrows():
            activity_tag = soup.new_tag('activity')
            activity_tag['type'] = activity['activity_type']
            
            if activity['facility']:
                activity_tag['facility'] = activity['facility']
            if activity['start_time']:
                activity_tag['start_time'] = activity['start_time']
            if activity['end_time']:
                activity_tag['end_time'] = activity['end_time']
            if activity['coordinates']:
                activity_tag['x'] = str(activity['coordinates'].x)
                activity_tag['y'] = str(activity['coordinates'].y)
            
            plan.append(activity_tag)
        
        root.append(person_tag)
    
    return str(soup)

class PopulationProcessor(BaseProcessor):
    def __init__(self, scales: Union[List[float], str, None] = None):
        super().__init__()
        self.scales = scales

    def process_chunk(self, chunk_file: Path) -> List[Path]:
        with open(chunk_file, 'r') as f:
            soup = BeautifulSoup(f, 'lxml-xml')
        
        population_data = process_population(soup)
        valid_scales = validate_scales(population_data.current_scale, self.scales)
        
        output_chunks = []
        for scale in valid_scales:
            output_chunk = self.temp_dir / f"population-{int(scale):02d}.xml"
            scaled_population = create_scaled_population(population_data, scale)
            output_xml = population_to_xml(scaled_population)
            
            with open(output_chunk, 'w') as f:
                f.write(output_xml)
            output_chunks.append(output_chunk)
        
        return output_chunks

def process_file(input_file: Path, output_dir: Path, scales: Union[List[float], str, None] = None):
    if not input_file.exists():
        raise FileNotFoundError(f"Input file not found: {input_file}")
    
    # Ensure population subdirectory exists
    population_dir = output_dir / "population"
    population_dir.mkdir(parents=True, exist_ok=True)
    
    processor = PopulationProcessor(scales)
    processor.process_file(input_file, population_dir)
