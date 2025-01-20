from setuptools import setup, find_packages

setup(
    name="ez-utils",
    version="0.1.0",
    packages=find_packages(where="src"),
    package_dir={"": "src"},
    install_requires=[
        "typer>=0.12.5",
        "rich>=13.7.1",
        "typing_extensions>=4.11.0",
        "pandas>=2.2.2",
        "numpy>=2.0.1",
        "beautifulsoup4>=4.12.3",
        "pyproj>=3.6.1",
        "shapely>=2.0.4",
        "lxml>=4.9.0"
    ],
    entry_points={
        "console_scripts": [
            "ez-utils=ez_utils.main:main",
        ],
    },
    author="k_nejads",
    description="CLI tool with 5 modules for various utilities",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    url="https://github.com/k_nejads/ez-utils",
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.6",
)
