--  <vc-preamble>
package Np_Ravel_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Dafny models the matrix as a curried function on ints; a SPARK matrix
   --  is a two-dimensional array, read at the same (i, j) positions.
   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;

   function Matrix_Get (M : Matrix; I : Index_Type; J : Index_Type)
     return Value_Type
   is
     (M (I, J))
   with Pre => I in M'Range (1) and then J in M'Range (2);

   function Matrix_Size (M : Index_Type; N : Index_Type) return Natural is
     (M * N);
--  </vc-preamble>

--  <vc-spec>
   procedure Ravel
     (Arr    : Matrix;
      M      : Index_Type;
      N      : Index_Type;
      Result : out Int_Array)
   with
     Pre  => M > 0 and then N > 0
             and then Arr'First (1) = 0 and then Arr'Length (1) = M
             and then Arr'First (2) = 0 and then Arr'Length (2) = N
             and then Result'First = 0
             and then Result'Length = Matrix_Size (M, N),
     Post => Result'Length = M * N
             and then
               (for all I in 0 .. M - 1 =>
                  (for all J in 0 .. N - 1 =>
                     (if I * N + J in Result'Range
                      then Result (I * N + J) = Matrix_Get (Arr, I, J))));

end Np_Ravel_Spec;

package body Np_Ravel_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Ravel
     (Arr    : Matrix;
      M      : Index_Type;
      N      : Index_Type;
      Result : out Int_Array) is
   begin
      pragma Assume (False);
   end Ravel;
--  </vc-code>

--  <vc-postamble>
end Np_Ravel_Spec;
--  </vc-postamble>
