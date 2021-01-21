from setuptools import setup, find_packages

setup(
    long_description=open("README.md", "r").read(),
    name="dhtc",
    version="0.42",
    description="dht crawler",
    author="Pascal Eberlein",
    author_email="pascal@eberlein.io",
    url="https://github.com/nbdy/dhtc",
    classifiers=[
        'Development Status :: 3 - Alpha',
        'Intended Audience :: Developers',
        'Topic :: Software Development :: Build Tools',
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python :: 3.6',
    ],
    keywords="dht crawler",
    packages=find_packages(),
    entry_points={
        'console_scripts': [
            'dhtc = dhtc.__main__:main'
        ]
    },
    package_data={
        "dhtc": ["templates/*.html"]
    },
    install_requires=open("requirements.txt").readlines()
)
