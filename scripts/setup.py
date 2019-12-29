from setuptools import setup

with open("Readme.md", 'r') as f:
    long_description = f.read()

setup(
   name='cleanup',
   version='1.0',
   description='Items used to cleanup after tests',
   license="MIT",
   long_description=long_description,
   packages=['cleaup'], 
   install_requires=requirements
)
