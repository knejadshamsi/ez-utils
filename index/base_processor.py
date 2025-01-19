import os
import shutil
from pathlib import Path
from typing import List, Generator, Union, Tuple
from abc import ABC, abstractmethod

class BaseProcessor(ABC):
    CHUNK_SIZE = 1000  # Lines per chunk

    def __init__(self):
        self.temp_dir = None

    def create_temp_dir(self) -> Path:
        if self.temp_dir:
            shutil.rmtree(self.temp_dir, ignore_errors=True)
        self.temp_dir = Path("temp") / self.__class__.__name__.lower()
        self.temp_dir.mkdir(parents=True, exist_ok=True)
        return self.temp_dir

    def chunk_file(self, input_file: Path) -> Generator[Path, None, None]:
        temp_dir = self.create_temp_dir()
        chunk_num = 0
        current_chunk = []
        
        with open(input_file, 'r', encoding='utf-8') as f:
            for line in f:
                current_chunk.append(line)
                if len(current_chunk) >= self.CHUNK_SIZE:
                    chunk_path = self._write_chunk(current_chunk, chunk_num)
                    yield chunk_path
                    chunk_num += 1
                    current_chunk = []
            
            if current_chunk:
                chunk_path = self._write_chunk(current_chunk, chunk_num)
                yield chunk_path

    def _write_chunk(self, lines: List[str], chunk_num: int) -> Path:
        chunk_path = self.temp_dir / f"chunk_{chunk_num}.txt"
        with open(chunk_path, 'w', encoding='utf-8') as f:
            f.writelines(lines)
        return chunk_path

    def merge_chunks(self, output_chunks: List[Path], output_file: Path):
        with open(output_file, 'w', encoding='utf-8') as outfile:
            for chunk in output_chunks:
                with open(chunk, 'r', encoding='utf-8') as infile:
                    shutil.copyfileobj(infile, outfile)

    def cleanup(self):
        if self.temp_dir and self.temp_dir.exists():
            shutil.rmtree(self.temp_dir)

    @abstractmethod
    def process_chunk(self, chunk_file: Path) -> Union[Path, Tuple[Path, ...]]:
        """Process a chunk file and return one or more output chunk paths"""
        pass

    def process_file(self, input_file: Path, *output_files: Path):
        """
        Process input file and write to one or more output files.
        
        Args:
            input_file: Path to input file to process
            *output_files: One or more output file paths
        """
        output_chunks_list = [[] for _ in output_files]
        try:
            for chunk_path in self.chunk_file(input_file):
                chunk_results = self.process_chunk(chunk_path)
                if isinstance(chunk_results, tuple):
                    for i, chunk in enumerate(chunk_results):
                        output_chunks_list[i].append(chunk)
                else:
                    output_chunks_list[0].append(chunk_results)
            
            for chunks, output_file in zip(output_chunks_list, output_files):
                self.merge_chunks(chunks, output_file)
        finally:
            self.cleanup()
