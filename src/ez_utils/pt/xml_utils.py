from bs4 import BeautifulSoup, Tag
from typing import Dict, Any, Optional

def create_xml_root(root_tag: str) -> BeautifulSoup:
    """Create a new XML document with root tag.
    
    Args:
        root_tag: Name of the root element
        
    Returns:
        BeautifulSoup object with XML document
    """
    return BeautifulSoup(f'<{root_tag}></{root_tag}>', 'lxml-xml')

def add_element(
    parent: Tag,
    name: str,
    attributes: Optional[Dict[str, Any]] = None,
    text: Optional[str] = None
) -> Tag:
    """Add a new element to parent with optional attributes and text.
    
    Args:
        parent: Parent BeautifulSoup Tag
        name: Element name
        attributes: Optional dict of attributes
        text: Optional text content
        
    Returns:
        Created BeautifulSoup Tag
    """
    soup = parent if isinstance(parent, BeautifulSoup) else parent.parent
    element = soup.new_tag(name)
    
    if attributes:
        for key, value in attributes.items():
            element[key] = str(value)
    
    if text is not None:
        element.string = str(text)
    
    parent.append(element)
    return element

def format_number(value: float, precision: int = 2) -> str:
    """Format number with specified precision.
    
    Args:
        value: Number to format
        precision: Number of decimal places
        
    Returns:
        Formatted string
    """
    return f"{value:.{precision}f}"
