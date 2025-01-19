from bs4 import BeautifulSoup
import pandas as pd
import numpy as np
from pathlib import Path
from typing import List, Dict, Tuple
from .models import Activity, Person, PopulationData, ScaledPopulation
from .utils import create_point_from_coordinates, parse_time, get_total_population
from .base_processor import BaseProcessor

def process_activities(person_soup: BeautifulSoup) -> Tuple[List[Activity], pd.Series]:
    activities = []
    activity_data = {}
    
    person_id = person_soup['id']
    plan = person_soup.find('plan')
    
    for i, activity in enumerate(plan.find_all('activity'), start=1):
        activity_obj = Activity(
            id=person_id,
            activity_order=i,
            activity_type=activity.get('type'),
            facility=activity.get('facility'),
            start_time=parse_time(activity.get('start_time')),
            end_time=parse_time(activity.get('end_time')),
            coordinates=create_point_from_coordinates(
                float(activity['x']), float(activity['y'])
            ) if 'x' in activity.attrs and 'y' in activity.attrs else None
        )
        activities.append(activity_obj)
        
        # Collect data for DataFrame
        activity_data.update({
            'id': person_id,
            'activity_order': i,
            'activity_type': activity_obj.activity_type,
            'facility': activity_obj.facility,
            'start_time': activity_obj.start_time,
            'end_time': activity_obj.end_time,
            'coordinates': activity_obj.coordinates
        })
    
    return activities, pd.Series(activity_data)

def process_population(population_soup: BeautifulSoup) -> PopulationData:
    persons = []
    activities_data = []
    
    for person in population_soup.find_all('person'):
        activities, activity_data = process_activities(person)
        persons.append(Person(id=person['id'], activities=activities))
        activities_data.append(activity_data)
    
    activities_df = pd.DataFrame(activities_data)
    total_population = get_total_population()
    current_scale = (len(persons) / total_population) * 100
    
    return PopulationData(
        persons=persons,
        activities_df=activities_df,
        total_population=total_population,
        current_scale=current_scale
    )

def create_scaled_population(population_data: PopulationData, target_scale: float) -> ScaledPopulation:
    scale_ratio = target_scale / population_data.current_scale
    keep_count = int(len(population_data.persons) * scale_ratio)
    
    selected_indices = np.random.choice(
        len(population_data.persons), 
        size=keep_count, 
        replace=False
    )
    
    selected_persons = [population_data.persons[i] for i in selected_indices]
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
    def process_chunk(self, chunk_file: Path) -> Path:
        output_chunk = self.temp_dir / f"output_{chunk_file.name}"
        
        with open(chunk_file, 'r') as f:
            soup = BeautifulSoup(f, 'lxml-xml')
        
        population_data = process_population(soup)
        scaled_population = create_scaled_population(population_data, target_scale=100)  # Default no scaling
        output_xml = population_to_xml(scaled_population)
        
        with open(output_chunk, 'w') as f:
            f.write(output_xml)
        
        return output_chunk

def process_file(input_file: Path, output_file: Path):
    processor = PopulationProcessor()
    processor.process_file(input_file, output_file)
